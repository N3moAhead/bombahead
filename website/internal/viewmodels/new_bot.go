package viewmodels

import (
	"regexp"
	"strings"

	"github.com/N3moAhead/bombahead/website/internal/models"
)

type NewBotForm struct {
	Name          string
	Description   string
	DockerHubUrl  string
	CreatedWithAi bool

	Errors map[string]string
}

func (f *NewBotForm) Validate() bool {
	f.Errors = make(map[string]string)

	if strings.TrimSpace(f.Name) == "" {
		f.Errors["Name"] = "Name is required."
	}
	if len(f.Name) > 30 {
		f.Errors["Name"] = "Name cannot be longer than 30 characters."
	}

	if len(f.Description) > 500 {
		f.Errors["Description"] = "Description cannot be longer than 500 characters."
	}

	if strings.TrimSpace(f.DockerHubUrl) == "" {
		f.Errors["DockerHubUrl"] = "Docker Hub URL is required."
	} else {
		// It must start with 'docker.io/' or 'ghcr.io/', have a user/namespace, and MUST have a tag.
		dockerImageRegex := `^(?:docker\.io|ghcr\.io)\/([a-zA-Z0-9.-]+\/)*[a-zA-Z0-9.-]+:[a-zA-Z0-9.-]+$`
		re := regexp.MustCompile(dockerImageRegex)

		if !re.MatchString(f.DockerHubUrl) {
			f.Errors["DockerHubUrl"] = "Please enter a valid Docker image URL starting with docker.io/ or ghcr.io/ and include a specific tag (e.g., ghcr.io/user/repo:1.0.0)."
		} else if strings.HasSuffix(f.DockerHubUrl, ":latest") {
			f.Errors["DockerHubUrl"] = "The 'latest' tag is not allowed. Please use a specific version tag."
		}
	}

	return len(f.Errors) == 0
}

func (n *NewBotForm) ToDbModel(userID uint) *models.Bot {
	return &models.Bot{
		Name:          n.Name,
		Description:   n.Description,
		DockerHubUrl:  n.DockerHubUrl,
		CreatedWithAi: n.CreatedWithAi,
		UserID:        userID,
	}
}
