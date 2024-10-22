# DO NOT EDIT THIS FILE MANUALLY! Use `release update-releases-file`.
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

CONFIG_LINUX_AMD64 = "linux-amd64"
CONFIG_LINUX_ARM64 = "linux-arm64"
CONFIG_DARWIN_AMD64 = "darwin-10.9-amd64"
CONFIG_DARWIN_ARM64 = "darwin-11.0-arm64"

_CONFIGS = [
    ("23.2.12", [
        (CONFIG_DARWIN_AMD64, "cbf3e43750edcc53dbb40bc93a84bbf12809d3cc3ca6478834c3b66dbb6d5741"),
        (CONFIG_DARWIN_ARM64, "851cb6e386332880e65c38d955cfec2760c0540570453011232f3bc0f163d9f4"),
        (CONFIG_LINUX_AMD64, "7e4b306a61931c3b61ce99a47f2a2cf5ab4f00c32cf2b72bd1d9af5d173e065c"),
        (CONFIG_LINUX_ARM64, "2ec1826dfae911355c87d8cb1d703566c85a746d6c60895169ceec924b528dc2"),
    ]),
    ("24.1.6", [
        (CONFIG_DARWIN_AMD64, "0d900af86357f5883ce10935bae7ea00e16ffc2d7875e56491e6d731f4565d9d"),
        (CONFIG_DARWIN_ARM64, "985e67e66bc29955f1547f7cc0748db5532ab0c57628bdf1ce3df3c9c1fc072a"),
        (CONFIG_LINUX_AMD64, "1120fae532f5e31411d8df06c9dac337b8116f1b167988ec2da675770c65a329"),
        (CONFIG_LINUX_ARM64, "9d913a9080bc777645aa8a6c009f717f500856f8b3b740d5bd9e8918ddd0d88a"),
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
