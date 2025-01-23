package claim

func in(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

func AuthStaff(staff string) bool {
	staffList := []string{"Jiankun", "Zhantang", "Fede"}
	result := in(staff, staffList)
	return result
}

func AuthManager(manager string) bool {
	managerList := []string{"manager1", "manager2", "manager3"}
	result := in(manager, managerList)
	return result
}
