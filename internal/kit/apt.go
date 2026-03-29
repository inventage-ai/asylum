package kit

func init() {
	Register(&Kit{
		Name:        "apt",
		Description: "Extra apt packages installed in the project image",
	})
}
