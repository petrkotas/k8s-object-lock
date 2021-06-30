// +build tools

package tools

import _ "k8s.io/code-generator" // so the codegenerator is vendored, otherwise it is ignored and codegen wont work
