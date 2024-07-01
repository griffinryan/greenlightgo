package graphics

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ObjData struct {
	Vertices []float32
	Colors   []float32
}

func LoadOBJ(filePath string) (ObjData, error) {
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
