package apiModels

import (
	"link-availability-checker/internal/models"
)

type Error struct {
	Error string `json:"error"`
}

type CheckLinkSetRequest struct {
	Links []string `json:"links" binding:"required"`
}

func (l *CheckLinkSetRequest) ConvertLinksToModel() []models.Link {
	result := make([]models.Link, 0, len(l.Links))
	for _, link := range l.Links {
		result = append(result, models.Link{Domain: link})
	}
	return result
}

type CheckLinkSetResponse struct {
	Links    map[string]string `json:"links"`
	LinksNum int               `json:"links_num"`
}

type GetLinkSetRequest struct {
	LinksList []int `json:"links_list" binding:"required"`
}

type GetLinkSetResponse struct {
	File []byte `json:"file"`
}
