package asmdb

type Instruction struct {
	Mnemonics []string
}

type Database struct {
	Instructions []*Instruction
}

type processor struct {
	raw *Raw
}

func Process(r *Raw) (*Database, error) {
	p := &processor{
		raw: r,
	}
	return p.process()
}

func (p *processor) process() (*Database, error) {
	db := &Database{}
	for _, i := range p.raw.Instructions {
		inst, err := p.instruction(i)
		if err != nil {
			return nil, err
		}
		db.Instructions = append(db.Instructions, inst)
	}
	return db, nil
}

func (p *processor) instruction(fields []string) (*Instruction, error) {
	return nil, nil
}
