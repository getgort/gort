package config

// bundles:
// - name: test
//   description: A test bundle.
//   docker:
//     image: clockworksoul/echotest
//     tag: latest
//   commands:
//     splitecho:
//       description: Echos back anything sent to it, one parameter at a time.
//       executable: /opt/app/splitecho.sh
//     curl:
//       description: Echos back anything sent to it, one parameter at a time.
//       executable: /home/bundle/tweet_cog_wrapper.sh
//     echo:
//       description: Echos back anything sent to it, one parameter at a time.
//       executable: /home/bundle/tweet_cog_wrapper.sh

type BundleConfig struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Docker      BundleDockerConfig `json:"docker"`
	// Commands    []BundleCommandConfig `json:"commands"`
}

// type BundleCommandConfig struct {
// 	Description string `json:"description"`
// 	Executable  string `json:"executable"`
// }

type BundleDockerConfig struct {
	Image string `json:"image"`
	Tag   string `json:"tag"`
}
