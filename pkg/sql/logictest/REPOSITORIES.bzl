# DO NOT EDIT THIS FILE MANUALLY! Use `release update-releases-file`.
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

CONFIG_LINUX_AMD64 = "linux-amd64"
CONFIG_LINUX_ARM64 = "linux-arm64"
CONFIG_DARWIN_AMD64 = "darwin-10.9-amd64"
CONFIG_DARWIN_ARM64 = "darwin-11.0-arm64"

_CONFIGS = [
    ("23.2.16", [
        (CONFIG_DARWIN_AMD64, "8568f117355358182beb7900b7cee34f1c9b7fb13fc5948a1b634cdc1ac1c378"),
        (CONFIG_DARWIN_ARM64, "22e80116189efbf46aee3cc227170794cdedd21a5587585bafdf3be739061266"),
        (CONFIG_LINUX_AMD64, "b504badee2e6dd2934d59462a6bcf6cffc77a997b58b2837ff09bfa181b7ae9d"),
        (CONFIG_LINUX_ARM64, "200723c91e81c948bd4cd45ef87d0a3f0ed9d34f9c21145c194f1a6bd680417b"),
    ]),
    ("24.1.7", [
        (CONFIG_DARWIN_AMD64, "da8494c0dc41546da460f3d338092ee06ee9873a7b345b0b548f22db01389d56"),
        (CONFIG_DARWIN_ARM64, "4eed1f37ff187230df0f298948f39d896f4868573df444c5ca07eab06b126b37"),
        (CONFIG_LINUX_AMD64, "07693f5df2f6704a4d33910962621d2190daffcf5082e9c24d11ca5470848c9c"),
        (CONFIG_LINUX_ARM64, "f57bfff82e9248e2780ee0dd4863faf9a5452622cb061dfbdb90260cb0a5c09f"),
    ]),
]

def _munge_name(s):
    return s.replace("-", "_").replace(".", "_")

def _repo_name(version, config_name):
    return "cockroach_binary_v{}_{}".format(
        _munge_name(version),
        _munge_name(config_name))

def _file_name(version, config_name):
    return "cockroach-v{}.{}/cockroach".format(
        version, config_name)

def target(config_name):
    targets = []
    for versionAndConfigs in _CONFIGS:
        version, _ = versionAndConfigs
        targets.append("@{}//:{}".format(_repo_name(version, config_name),
                                         _file_name(version, config_name)))
    return targets

def cockroach_binaries_for_testing():
    for versionAndConfigs in _CONFIGS:
        version, configs = versionAndConfigs
        for config in configs:
            config_name, shasum = config
            file_name = _file_name(version, config_name)
            http_archive(
                name = _repo_name(version, config_name),
                build_file_content = """exports_files(["{}"])""".format(file_name),
                sha256 = shasum,
                urls = [
                    "https://binaries.cockroachdb.com/{}".format(
                        file_name.removesuffix("/cockroach")) + ".tgz",
                ],
            )
