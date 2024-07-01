#version 410

layout(location = 0) in vec3 vert;
layout(location = 1) in vec3 vertColor;

out vec3 fragColor;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

void main() {
    fragColor = vertColor;
    gl_Position = projection * view * model * vec4(vert, 1.0);
}
