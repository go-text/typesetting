package fontscan

func init() {
	// replace families keys by their no case no blank version
	for i, v := range familySubstitution {
		for i, s := range v.additionalFamilies {
			v.additionalFamilies[i] = ignoreBlanksAndCase(s)
		}

		familySubstitution[i].targetFamily = ignoreBlanksAndCase(v.targetFamily)
	}
}

// familySubstitution maps family name to possible alias
// it is generated from fontconfig substitution rules
// the order matters, since the rules apply to the current
// state of the family list
// TODO: generate the list
var familySubstitution = []substitution{}
