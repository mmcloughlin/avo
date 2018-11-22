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
	Encoding         Encoding          `xml:"Encoding"`
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

type Encoding struct {
	REX  *REX  `xml:"REX"`
	VEX  *VEX  `xml:"VEX"`
	EVEX *EVEX `xml:"EVEX"`
}

type REX struct {
	Mandatory bool   `xml:"mandatory,attr"`
	W         int    `xml:"W,attr"`
	R         string `xml:"R,attr"`
	X         string `xml:"X,attr"`
	B         string `xml:"B,attr"`
}

type VEX struct {
	Type string `xml:"type,attr"`
	W    int    `xml:"W,attr"`
	L    int    `xml:"L,attr"`
	M5   string `xml:"m-mmmm,attr"`
	PP   string `xml:"pp,attr"`
	R    string `xml:"R,attr"`
	X    string `xml:"X,attr"`
	B    string `xml:"B,attr"`
	V4   string `xml:"vvvv,attr"`
}

type EVEX struct {
	M2      string `xml:"mm,attr"`
	PP      string `xml:"pp,attr"`
	W       int    `xml:"W,attr"`
	LL      string `xml:"LL,attr"`
	V4      string `xml:"vvvv,attr"`
	V       string `xml:"V,attr"`
	RR      string `xml:"RR,attr"`
	B       string `xml:"B,attr"`
	X       string `xml:"X,attr"`
	Bsml    string `xml:"b,attr"`
	A3      string `xml:"aaa,attr"`
	Z       string `xml:"Z,attr"`
	Disp8xN string `xml:"disp8xN,attr"`
}
