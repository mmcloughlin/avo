package printer

// ARM64WritesNZCVForTest exposes the flag-effect classification to tests.
func ARM64WritesNZCVForTest(mnemonic string) (writes, known bool) {
	w, k := arm64WritesNZCV[mnemonic]
	return w, k
}
