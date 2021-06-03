Contains:

* adapter: An interface representing an Adapter, which connects to a chat provider and converts provider-specific events into Gort events.
* events: descriptions of the various Gort events, provided by an Adapter implementation
* tokenize: splits a command string into parameter tokens.

Also contains some data wrapper structs:

* ProviderInfo - info about a generic provider
* ChannelInfo - info about a generic provider "channel" (which can contain provider-specific values, like ID)
* UserInfo - info about a generic user (which can contain provider-specific values, like ID)
