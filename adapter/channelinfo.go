package adapter

// ChannelInfo contains the basic information for a single channel in any provider.
type ChannelInfo struct {
	ID      string
	Members []string
	Name    string
}
