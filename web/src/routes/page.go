package routes

type Page struct {
	Number int
	Offset int
}

func GetPages(total, limit int64) (pages []*Page) {
	if total <= limit {
		return
	}
	number := 0
	for start := 0; start < int(total); start += int(limit) {
		number++
		page := &Page{
			Number: number,
			Offset: start,
		}
		pages = append(pages, page)
	}
	return
}
