package esr

var C_ddr_str [4]byte

func (ESRM *ESRManager) Fn_init_ddr_pop(c_PipeID string, c_Origin string, c_PrdctTyp string) error {
	if c_PipeID[0] == '1' {
		// Copy the second character of cPipeID into the first three positions of cDdrStr
		C_ddr_str[0] = c_PipeID[1]
		C_ddr_str[1] = c_PipeID[1]
		C_ddr_str[2] = c_PipeID[1]
		C_ddr_str[3] = 0 // Null equivalent in byte array (to mimic C behavior)
	} else {
		// Copy "000" into cDdrStr
		copy(C_ddr_str[:], "000")
		C_ddr_str[3] = 0 // Null equivalent
	}

	return nil
}

func (ESRM *ESRManager) Fn_cpy_ddr(c_ddr_var *[4]byte) error {

	*c_ddr_var = C_ddr_str

	return nil
}
