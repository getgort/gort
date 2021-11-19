# Use `tilt up -- --k8s` to run with Kubernetes.
config.define_bool("k8s",args=False,usage="If set, runs resources under Kubernetes. Otherwise, docker-compose is used.")

args = config.parse()
k8s = "k8s" in args and args["k8s"]

# Build the container image as needed
docker_build('getgort/gort', '.')

# Set up resources using docker-compose
def compose():
    docker_compose('docker-compose.yml')

# Pull relay config from development.yml
def loadRelayConfig(devConfig,relayType,setValues):
    if relayType in devConfig.keys():
        for i in range(len(devConfig[relayType])):
            discord = devConfig[relayType][i]
            for key in discord:
                setValues.append("config.{0}[{1}].{2}={3}".format(relayType,i,key,discord[key]))

def loadDbConfig(devConfig,setValues):
    db = devConfig["database"]
    for key in db:
        value = "config.{0}.{1}={2}".format("database",key,db[key])
        setValues.append(value)

setValues = []
if os.path.exists("development.yml"):
    devConfig = read_yaml("development.yml")
    loadRelayConfig(devConfig,"discord",setValues)
    loadRelayConfig(devConfig,"slack",setValues)
    loadDbConfig(devConfig,setValues)

# Set up resources using Kubernetes
def kubernetes():
    k8s_yaml('tilt-datasources.yaml')
    k8s_yaml(
        helm(
            './helm/gort',
            set = setValues
        )
    )
    k8s_resource('postgres', port_forwards=5432)
    k8s_resource('chart-gort', port_forwards=4000)

## Resources to permit common tasks to be run via the Tilt Web UI.

# Bootstrapping the server
local_resource(
    "Bootstrap",
    "go run . bootstrap https://localhost:4000 --allow-insecure",
    trigger_mode=TRIGGER_MODE_MANUAL,
    auto_init=False
)

# Clear profiles
local_resource(
    "Clear Profiles",
    "rm -f ~/.gort/profile",
    trigger_mode=TRIGGER_MODE_MANUAL,
    auto_init=False
)

if k8s:
    kubernetes()
else:
    compose()
