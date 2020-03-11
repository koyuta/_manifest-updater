package registry

type Registry interface {
	FetchLatestTag() (string, error)
}
