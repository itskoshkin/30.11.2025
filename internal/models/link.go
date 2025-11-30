package models

type Link struct {
	Domain string
	Status bool
}

type Set struct {
	Number int
	Links  []Link
}

func (s *Set) ConvertLinksToStrMap() map[string]string {
	result := make(map[string]string, len(s.Links))
	for _, link := range s.Links {
		//if link.Status {
		//	result[link.Domain] = "available"
		//} else {
		//	result[link.Domain] = "not available"
		//}
		result[link.Domain] = ConvertStatusToString(link.Status)
	}
	return result
}

func ConvertStatusToString(status bool) string {
	if status {
		return "available"
	} else {
		return "not available"
	}
}
