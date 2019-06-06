package module

//LoadModulesList ..
var LoadModulesList = []Generator{}

//RegModule ..
func RegModule(g Generator) {
	LoadModulesList = append(LoadModulesList, g)
}
