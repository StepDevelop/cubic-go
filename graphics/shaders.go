package graphics

import (
    "github.com/go-gl/gl/v3.3-core/gl"
    "strings"
)

const cameraFragShader string = `
#version 330 core
uniform sampler2DArray texArray;
in vec3 texCoord;
void main() {
    gl_FragColor = texture(texArray, texCoord);
}` + "\x00"

const cameraVertShader string = `
#version 330 core
uniform mat4 perspectiveMatrix; // perspective matrix (constant)
uniform mat4 translationMatrix; // translation matrix (varies with camera state)
uniform mat4 rotXMatrix; // x-rotation matrix (varies with camera state)
uniform mat4 rotYMatrix; // y-rotation matrix (varies with camera state)
uniform mat4 rotZMatrix; // z-rotation matrix (varies with camera state)
out vec3 texCoord;
attribute vec4 in_position;
attribute vec3 in_texCoord;
void main() {
    gl_Position = perspectiveMatrix * rotXMatrix * rotYMatrix * rotZMatrix * translationMatrix * in_position;
    texCoord = in_texCoord;
}` + "\x00"

var CameraShader uint32

func initCameraShader() {
    CameraShader = createProgram(cameraVertShader, cameraFragShader)
}

func createProgram(vert, frag string) uint32 {
    vertHandle := compileShader(gl.VERTEX_SHADER, &vert)
    fragHandle := compileShader(gl.FRAGMENT_SHADER, &frag)

    return linkProgram(vertHandle, fragHandle)
}

func compileShader(shaderType uint32, source *string) uint32 {
    handle := gl.CreateShader(shaderType)
    csources, free := gl.Strs(*source)
    gl.ShaderSource(handle, 1, csources, nil)
    free()

    gl.CompileShader(handle)

    var result int32
    gl.GetShaderiv(handle, gl.COMPILE_STATUS, &result)
    if result == gl.FALSE {
        panic(*getShaderLog(handle, false))
    }

    return handle
}

func linkProgram(vertHandle, fragHandle uint32) uint32 {
    var programHandle uint32 = gl.CreateProgram()
    gl.AttachShader(programHandle, vertHandle)
    gl.AttachShader(programHandle, fragHandle)
    gl.LinkProgram(programHandle)

    var result int32
    gl.GetProgramiv(programHandle, gl.LINK_STATUS, &result)
    if result == gl.FALSE {
        panic(*getShaderLog(programHandle, true))
    }

    gl.DeleteShader(vertHandle)
    gl.DeleteShader(fragHandle)

    return programHandle
}

func getShaderLog(handle uint32, program bool) *string {
    var lenFunc func(uint32, uint32, *int32)
    var logFunc func(uint32, int32, *int32, *uint8)
    if program {
        lenFunc = gl.GetProgramiv
        logFunc = gl.GetProgramInfoLog
    } else {
        lenFunc = gl.GetShaderiv
        logFunc = gl.GetShaderInfoLog
    }

    var maxLen int32
    lenFunc(handle, gl.INFO_LOG_LENGTH, &maxLen)
    log := strings.Repeat("\x00", int(maxLen + 1))
    logFunc(handle, maxLen, nil, gl.Str(log))
    return &log
}
