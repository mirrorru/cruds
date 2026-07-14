package samples

type IdNameSmpl struct {
	ID   int    `tbl:"pk"`
	Name string `tbl:"sort=1"`
}

type IdNameAgeSmpl struct {
	IdNameSmpl
	Age *int
}

type TwoKey struct {
	Key1 int    `tbl:"pk"`
	Key2 string `tbl:"pk"`
}

type TwoFields struct {
	Fld1 int    `tbl:"col=field_one"`
	Fld2 string `tbl:"col=field_two"`
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
	TwoFlds TwoFields `tbl:"embd"` // NOT expands to next two fields
}
