package main

type TForward struct {
	state *TState
	path  string
	ast   *TAst
}

func (f *TForward) forwardStruct(str *TAst) {

}

func (f *TForward) forwardFunc(fn *TAst) {

}

func (f *TForward) forwardImport(imp *TAst) {

}

func (f *TForward) forward() {
	body := f.ast.astArr0
	for i := 0; i < len(body); i++ {
		child := body[i]
		switch child.Ttype {
		case AstStruct:
			f.forwardStruct(child)
		case AstFunc:
			f.forwardFunc(child)
		case AstImport:
			f.forwardImport(child)
		}
	}
}

func forwardDeclairation(state *TState, path string, ast *TAst) {
	// Create a new forward declaration
	forward := &TForward{
		state: state,
		path:  path,
		ast:   ast,
	}
	forward.forward()
}
