package packet

var facilities = []string{"yyz1", "nrt1", "atl1", "mrs1", "hkg1", "ams1", "ewr1", "sin1", "dfw1", "lax1", "syd1", "sjc1", "ord1", "iad1", "fra1", "sea1"}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
