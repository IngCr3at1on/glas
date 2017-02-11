package core

var (
	// Directions defines a list of standard directions.
	Directions = []string{"n", "s", "e", "w", "ne", "nw", "se", "sw", "u", "d"}

	churchToSewers = []string{"s", "e", "n", "d"}
	churchToGuild  = []string{"s", "e", "e", "s"}
	churchToShop   = []string{"s", "e", "e", "n"}

	sewersToChurch = []string{"u", "s", "w", "n"}
	sewersToGuild  = []string{"u", "s", "e", "s"}
	sewersToShop   = []string{"u", "s", "e", "n"}

	shopToChurch = []string{"s", "w", "w", "n"}
	shopToGuild  = []string{"s", "s"}
	shopToSewers = []string{"s", "w", "n", "d"}
)

func goToSewers() []string {
	return churchToSewers
}
