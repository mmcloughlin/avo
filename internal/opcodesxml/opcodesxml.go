package opcodesxml

import (
	"encoding/xml"
	"io"
	"os"
)

func Read(r io.Reader) (*InstructionSet, error) {
	d := xml.NewDecoder(r)
	is := &InstructionSet{}
	if err := d.Decode(is); err != nil {
		return nil, err
	}
	return is, nil
}

func ReadFile(filename string) (*InstructionSet, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Read(f)
}

type InstructionSet struct {
	Name         string        `xml:"name,attr"`
	Instructions []Instruction `xml:"Instruction"`
}

type Instruction struct {
	Name    string `xml:"name,attr"`
	Summary string `xml:"summary,attr"`
	Forms   []Form `xml:"InstructionForm"`
}

type Form struct {
	GASName          string            `xml:"gas-name,attr"`
	GoName           string            `xml:"go-name,attr"`
	MMXMode          string            `xml:"mmx-mode,attr"`
	XMMMode          string            `xml:"xmm-mode,attr"`
	CancellingInputs bool              `xml:"cancelling-inputs,attr"`
	Operands         []Operand         `xml:"Operand"`
	ImplicitOperands []ImplicitOperand `xml:"ImplicitOperand"`
	ISA              []ISA             `xml:"ISA"`
}

type Operand struct {
	Type   string `xml:"type,attr"`
	Input  bool   `xml:"input,attr"`
	Output bool   `xml:"output,attr"`
}

type ImplicitOperand struct {
	ID     string `xml:"id,attr"`
	Input  bool   `xml:"input,attr"`
	Output bool   `xml:"output,attr"`
}

type ISA struct {
	ID string `xml:"id,attr"`
}
