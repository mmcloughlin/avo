package preproc

import "github.com/mmcloughlin/avo/build"

// Ifdef inserts an #ifdef preprocessor macro. Condition should be a valid
// preprocessor expression, though this is not checked.
func Ifdef(condition string) {
	build.GlobalContext().PreprocIfdef(condition)
}

// Ifndef inserts an #ifndef preprocessor macro. Condition should be a valid
// preprocessor expression, though this is not checked.
func Ifndef(condition string) {
	build.GlobalContext().PreprocIfndef(condition)
}

// Else inserts an #else preprocessor macro. It needs to be preceded by an
// #ifdef or #ifndef macro.
func Else() {
	build.GlobalContext().PreprocElse()
}

// End inserts an #endif preprocessor macro. It needs to be preceded by an
// #ifdef, #ifndef, or #else macro.
func Endif() {
	build.GlobalContext().PreprocEndif()
}
