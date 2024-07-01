package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width  = 800
	height = 600
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

type ObjData struct {
	Vertices []float32
	Colors   []float32
}

func loadOBJ(filePath string) (ObjData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return ObjData{}, err
	}
	defer file.Close()

	var vertices []float32
	var colors []float32

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "v ") {
			var x, y, z float32
			fmt.Sscanf(line, "v %f %f %f", &x, &y, &z)
			vertices = append(vertices, x, y, z)
			// Adding color based on position
			colors = append(colors, (x+1)/2, (y+1)/2, (z+1)/2)
		}
	}

	if err := scanner.Err(); err != nil {
		return ObjData{}, err
	}

	var interleaved []float32
	for i := 0; i < len(vertices); i += 3 {
		interleaved = append(interleaved, vertices[i], vertices[i+1], vertices[i+2])
		interleaved = append(interleaved, colors[i], colors[i+1], colors[i+2])
	}

	return ObjData{Vertices: interleaved}, nil
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "3D Model Viewer", nil, nil)
	if err != nil {
		log.Fatalln("failed to create glfw window:", err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatalln("failed to initialize gl:", err)
	}

	gl.Enable(gl.DEPTH_TEST)

	// Load the OBJ model
	objData, err := loadOBJ("data/obj/miata.obj")
	if err != nil {
		log.Fatalln("failed to load obj model:", err)
	}

	// Create and compile the vertex and fragment shaders
	vertexShaderSource, err := ioutil.ReadFile("shaders/vertex_shader.glsl")
	if err != nil {
		log.Fatalln("failed to read vertex shader:", err)
	}
	vertexShader, err := compileShader(string(vertexShaderSource), gl.VERTEX_SHADER)
	if err != nil {
		log.Fatalln(err)
	}

	fragmentShaderSource, err := ioutil.ReadFile("shaders/fragment_shader.glsl")
	if err != nil {
		log.Fatalln("failed to read fragment shader:", err)
	}
	fragmentShader, err := compileShader(string(fragmentShaderSource), gl.FRAGMENT_SHADER)
	if err != nil {
		log.Fatalln(err)
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var linked int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &linked)
	if linked == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		logMsg := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(logMsg))

		log.Fatalln("failed to link program:", logMsg)
	}

	gl.UseProgram(program)

	// Create VBO and VAO
	var vbo, vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(objData.Vertices)*4, gl.Ptr(objData.Vertices), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))

	colorAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertColor\x00")))
	gl.EnableVertexAttribArray(colorAttrib)
	gl.VertexAttribPointer(colorAttrib, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))

	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	viewUniform := gl.GetUniformLocation(program, gl.Str("view\x00"))
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))

	for !window.ShouldClose() {
		// Clear the screen
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Calculate transformations
		model := mgl32.HomogRotate3DY(float32(glfw.GetTime()))
		view := mgl32.LookAtV(mgl32.Vec3{0, 0, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
		projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(width)/height, 0.1, 10.0)

		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
		gl.UniformMatrix4fv(viewUniform, 1, false, &view[0])
		gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

		// Draw the model
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(objData.Vertices)/6))

		// Swap buffers
		window.SwapBuffers()
		glfw.PollEvents()

		// Limit to 60 FPS
		time.Sleep(time.Second / 60)
	}
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		logMsg := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(logMsg))

		return 0, fmt.Errorf("failed to compile %v: %v", source, logMsg)
	}

	return shader, nil
}
