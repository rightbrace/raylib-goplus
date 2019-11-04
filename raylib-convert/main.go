package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {
	file, err := os.Open("headers.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	failed := make([]string, 0)
	prototypes := make([]*prototype, 0)
	success := make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		p, err := parseLine(line)
		if err == nil {
			if p != nil {
				prototypes = append(prototypes, p)
				trans, terr := translatePrototype(p)
				if terr == nil {
					success = append(success, trans)
				} else {
					fmt.Println("Failed: ", line, terr)
					failed = append(failed, "\n//"+terr.Error()+"\n"+line)
				}
			}
		} else {
			fmt.Println("Failed: ", line, err)
			failed = append(failed, "\n//"+err.Error()+"\n"+line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	//WRite the headers and the failures
	ioutil.WriteFile("out/headers.go", []byte("package raylib\n"+strings.Join(success, "\n")), 0644)
	ioutil.WriteFile("out/headers.fail.txt", []byte(strings.Join(failed, "\n")), 0644)
	fmt.Println("Completed ", len(success), " / ", len(prototypes), " functions")
	fmt.Println("DOES NOT HAVE RETURN TYPES YET")
}

func translatePrototype(prototype *prototype) (string, error) {

	//We have a manual definition, so use that instead
	if _, err := os.Stat("manual/" + prototype.name + ".go"); err == nil {
		bt, fe := ioutil.ReadFile("manual/" + prototype.name + ".go")
		return "\n" + string(bt), fe
	}

	//We do not support return types really yet, but when we do we have a special case for pointers
	if prototype.pointer && prototype.returnType != "char *" {
		return "", errors.New("cannot process pointer return types")
	}

	//Prepare some variables
	argHeaders := make([]string, len(prototype.args))
	bodyArgs := make([]string, len(prototype.args))
	bodyArgsTally := 0
	returnHeaders := make([]string, 1)
	body := "C." + prototype.name + "("

	//Convert the arguments into their headers
	for i, arg := range prototype.args {
		if arg == nil {
			continue
		}

		//Make sure it's a valid type
		if arg.pointer && arg.valueType != "char" {
			return "", errors.New("cannot process pointer arg types")
		}

		//Append to the header
		argHeaders[i] = arg.name + " " + convertType(arg.valueType)
		bodyArgPart, bodyPrefixPart := castType(*arg)

		//Append to C function header
		bodyArgs[bodyArgsTally] = bodyArgPart
		bodyArgsTally++

		//If we have a definition, then prepend it to the body
		if len(bodyPrefixPart) > 0 {
			body = bodyPrefixPart + "\n" + body
		}

	}

	//Finish the body and add everythign back
	body = body + strings.Join(bodyArgs, ", ") + ")"

	//Prepare the function
	text := "func " + prototype.name + "(" + strings.Join(argHeaders, ", ") + ") (" + strings.Join(returnHeaders, ", ") + ") {\n" + body + "\n}"
	return "//" + prototype.name + " : " + prototype.comment + "\n" + text, nil
}

//castType creates a cast for a type, returning first the name of the variable and then the definition of the variable.
// There are some cases where there is no definition.
func castType(a argument) (string, string) {
	csname := "c" + a.name

	switch a.valueType {
	default:
		deref := "*"
		if a.pointer {
			deref = ""
		}
		return csname, csname + " := " + deref + a.name + ".cptr()"
	case "float":
		fallthrough
	case "int":
		fallthrough
	case "uint8":
		fallthrough
	case "bool":
		return "C." + a.valueType + "(" + a.name + ")", ""
	case "char":
		return csname, csname + " := C.CString(" + a.name + ")\ndefer C.free(unsafe.Pointer(&" + csname + "))"
	}
}

//convertType converts a c type to a go type
func convertType(t string) string {
	switch t {
	default:
		return t
	case "float":
		return "float32"
	case "char":
		return "string"
	}
}

//Parses a line and generates a prototype
func parseLine(line string) (*prototype, error) {
	//Trim the line and validate it
	line = strings.Trim(line, " ")
	if len(line) < 4 || strings.HasPrefix(line, "//") {
		return nil, nil
		//return nil, errors.New("line is a comment or blank")
	}

	//rePrototype := regexp.MustCompile(`(RLAPI|RAYGUIDEF)\s+(\w{2,})\s+(\w+)\s?\(([^!@#$+%^]+?)\);\s*\/\/(.*)`)
	rePrototype := regexp.MustCompile(`(RLAPI)( const)?\s+([a-zA-Z0-9]{2,}(\s?\*)?)\s?(\w+)\s?\(([^!@#$+%^]+?)\);\s*\/\/(.*)`)
	reArgument := regexp.MustCompile(`(const |unsigned )?([a-zA-Z0-9]+) (\*?)([a-zA-Z0-9]+)`)

	matches := rePrototype.FindAllStringSubmatch(line, -1)
	if len(matches) != 1 {
		return nil, errors.New("invalid amount of matches for header")
	}

	//Prepare the prototype
	p := &prototype{
		entire:     matches[0][0],
		returnType: matches[0][3],
		pointer:    len(matches[0][4]) > 0,
		name:       matches[0][5],
		comment:    matches[0][7],
	}

	//Prepare the arguments
	parts := strings.Split(matches[0][6], ",")
	arguments := make([]*argument, len(parts))
	i := 0
	for _, p := range parts {
		matches := reArgument.FindAllStringSubmatch(p, -1)

		if len(matches) != 1 {
			if p == "void" {
				break
			} else {
				return nil, errors.New("invalid amount of matches for arguments")
			}
		}

		name := matches[0][4]
		if name == "type" || name == "interface" || name == "return" {
			name = "g" + name
		}

		arguments[i] = &argument{
			entire:    matches[0][0],
			valueType: matches[0][2],
			pointer:   len(matches[0][3]) > 0,
			name:      name,
		}

		i++
	}

	p.args = arguments
	return p, nil
}

type prototype struct {
	entire     string
	returnType string
	pointer    bool
	name       string
	args       []*argument
	comment    string
}

type argument struct {
	entire    string
	valueType string
	name      string
	pointer   bool
}
