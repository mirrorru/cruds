package samples

type IdNameSmpl struct {
	ID   int    `crud:"pk"`
	Name string `crud:"sort=1"`
}

type IdNameAgeSmpl struct {
	IdNameSmpl
	Age *int
}

type TwoKey struct {
	Key1 int    `crud:"pk"`
	Key2 string `crud:"pk"`
}

type TwoFields struct {
	Fld1 int    `crud:"col=field_one"`
	Fld2 string `crud:"col=field_two"`
}

type TwoCombo1Smpl struct {
	TwoKey    // expands to two fields
	TwoFields // expands to next two fields
}

func (TwoCombo1Smpl) SQLName() string {
	return "sample_table"
}

type TwoCombo2Smpl struct {
	TwoKey            // expands to two fields
	TwoFlds TwoFields // NOT expands to next two fields
}

type TwoCombo3Smpl struct {
	TwoKey            // expands to two fields
	TwoFlds TwoFields `crud:"embd"` // NOT expands to next two fields
}
