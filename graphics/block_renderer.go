package graphics

import (
    "github.com/go-gl/gl/v3.3-core/gl"
    "github.com/caseif/cubic-go/world"
    "github.com/go-gl/mathgl/mgl32"
)

const unitLength float32 = 0.5
const floatsPerVertex = 3
const floatsPerFace int = 6 * floatsPerVertex

var positionAttrIndex uint32 = 1<<31
var colorAttrIndex uint32 = 1<<31
var texCoordAttrIndex uint32 = 1<<31

var chunkVboHandles = make(map[*world.Chunk]uint32)
var chunkVboSizes = make(map[*world.Chunk]int32)
var chunkVaoHandles = make(map[*world.Chunk]uint32)

func render(world *world.World) {
    gl.UseProgram(CameraShader)
    gl.UniformMatrix4fv(gl.GetUniformLocation(CameraShader, gl.Str("trMatrix\x00")), 1, false,
        &CAMERA.GetTranslationMatrix()[0])
    gl.UniformMatrix4fv(gl.GetUniformLocation(CameraShader, gl.Str("rXMatrix\x00")), 1, false,
        &CAMERA.GetXRotationMatrix()[0])
    gl.UniformMatrix4fv(gl.GetUniformLocation(CameraShader, gl.Str("rYMatrix\x00")), 1, false,
        &CAMERA.GetYRotationMatrix()[0])
    gl.UniformMatrix4fv(gl.GetUniformLocation(CameraShader, gl.Str("rZMatrix\x00")), 1, false,
        &CAMERA.GetZRotationMatrix()[0])

    for _, chunk := range world.ChunkMap {
        if chunk.Dirty {
            vbo := createVbo(chunk)

            if handle := chunkVaoHandles[chunk]; handle != 0 {
                gl.DeleteVertexArrays(1, &handle)
            }

            chunkVboSizes[chunk] = int32(len(*vbo))
            handle, hasHandle := chunkVboHandles[chunk]
            if !hasHandle {
                gl.GenBuffers(1, &handle)
                chunkVboHandles[chunk] = handle
            }
            chunkVaoHandles[chunk] = prepareVbo(handle, vbo)

            chunk.Dirty = false
        }
        renderChunk(chunk)
    }
}

func createVbo(chunk *world.Chunk) *[]float32 {
    var buffer []float32

    for x := 0; x < world.CHUNK_LENGTH; x++ {
        for y := 0; y < world.CHUNK_LENGTH; y++ {
            for z := 0; z < world.CHUNK_LENGTH; z++ {
                var b *world.Block = chunk.Blocks[x][y][z]
                if b == nil {
                    continue
                }
                blockType := b.GetType()

                rX := float32(x) * unitLength
                rY := float32(y) * unitLength
                rZ := float32(z) * unitLength

                faces := make([][floatsPerFace]float32, 0, 6)

                c000 := &mgl32.Vec3{rX,              rY,              rZ}
                c001 := &mgl32.Vec3{rX,              rY,              rZ + unitLength}
                c010 := &mgl32.Vec3{rX,              rY + unitLength, rZ}
                c011 := &mgl32.Vec3{rX,              rY + unitLength, rZ + unitLength}
                c100 := &mgl32.Vec3{rX + unitLength, rY,              rZ}
                c101 := &mgl32.Vec3{rX + unitLength, rY,              rZ + unitLength}
                c110 := &mgl32.Vec3{rX + unitLength, rY + unitLength, rZ}
                c111 := &mgl32.Vec3{rX + unitLength, rY + unitLength, rZ + unitLength}

                // back face
                if b.GetRelative(world.BACK) == nil {
                    faces = append(faces, *createQuad(blockType, world.BACK, c100, c000, c010, c110))
                }
                // front face
                if b.GetRelative(world.FRONT) == nil {
                    faces = append(faces, *createQuad(blockType, world.FRONT, c001, c101, c111, c011))
                }
                // left face
                if b.GetRelative(world.LEFT) == nil {
                    faces = append(faces, *createQuad(blockType, world.LEFT, c000, c001, c011, c010))
                }
                // right face
                if b.GetRelative(world.RIGHT) == nil {
                    faces = append(faces, *createQuad(blockType, world.RIGHT, c101, c100, c110, c111))
                }
                // bottom face
                if b.GetRelative(world.BOTTOM) == nil {
                    faces = append(faces, *createQuad(blockType, world.BOTTOM, c000, c100, c101, c001))
                }
                // top face
                if b.GetRelative(world.TOP) == nil {
                    faces = append(faces, *createQuad(blockType, world.TOP, c010, c011, c111, c110))
                }

                for _, face := range faces {
                    buffer = append(buffer, face[:]...)
                }
            }
        }
    }

    return &buffer
}

func createQuad(blockType world.BlockType, face world.BlockFace, v0, v1, v2, v3 *mgl32.Vec3) *[floatsPerFace]float32 {
    var buffer [floatsPerFace]float32
    var bSlice = buffer[0:0]
    createVertex(&bSlice, v0, blockType, face, 0)
    createVertex(&bSlice, v1, blockType, face, 1)
    createVertex(&bSlice, v2, blockType, face, 2)
    createVertex(&bSlice, v0, blockType, face, 0)
    createVertex(&bSlice, v2, blockType, face, 2)
    createVertex(&bSlice, v3, blockType, face, 3)
    return &buffer
}

func createVertex(buffer *[]float32, location *mgl32.Vec3, blockType world.BlockType, face world.BlockFace,
ordinal int) {
    *buffer = append(*buffer, location.X())
    *buffer = append(*buffer, location.Y())
    *buffer = append(*buffer, location.Z())
    //TODO: textures and shit
}

func prepareVbo(handle uint32, vbo *[]float32) uint32 {
    var vaoHandle uint32
    gl.GenVertexArrays(1, &vaoHandle)
    gl.BindVertexArray(vaoHandle)
    gl.BindBuffer(gl.ARRAY_BUFFER, handle)

    gl.BufferData(gl.ARRAY_BUFFER, len(*vbo) * 4, gl.Ptr(*vbo), gl.STATIC_DRAW)

    checkIndices()

    gl.EnableVertexAttribArray(positionAttrIndex)
    //gl.EnableVertexAttribArray(texCoordAttrIndex)

    gl.VertexAttribPointer(positionAttrIndex, 3, gl.FLOAT, false, 12, nil)
    //gl.VertexAttribPointer(texCoordAttrIndex, 3, gl.FLOAT, false, 24, nil)

    gl.BindVertexArray(0)

    return vaoHandle
}

func renderChunk(chunk *world.Chunk) {
    gl.BindVertexArray(chunkVaoHandles[chunk])
    gl.DrawArrays(gl.TRIANGLES, 0, chunkVboSizes[chunk] / floatsPerVertex)
    gl.BindVertexArray(0)
}

func checkIndices() {
    if positionAttrIndex == 1<<31 {
        positionAttrIndex = uint32(gl.GetAttribLocation(CameraShader, gl.Str("in_position\x00")))
    }
    if colorAttrIndex == 1<<31 {
        colorAttrIndex = uint32(gl.GetAttribLocation(CameraShader, gl.Str("in_color\x00")))
    }
    if texCoordAttrIndex == 1<<31 {
        texCoordAttrIndex = uint32(gl.GetAttribLocation(CameraShader, gl.Str("in_texCoord\x00")))
    }
}
