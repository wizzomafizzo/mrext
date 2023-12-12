#!/usr/bin/env python3

import yaml


def extract_platform_ids(yaml_data):
    platform_ids = set()

    for section in yaml_data:
        if "cores" in section:
            for core in section["cores"]:
                if "platform_id" in core:
                    platform_ids.add(core["platform_id"])

    return sorted(list(platform_ids))


def main():
    file_path = "cores.yml"

    with open(file_path, "r") as file:
        yaml_data = yaml.safe_load(file)

    if not yaml_data:
        print("YAML file is empty or invalid")
        return

    platform_ids = extract_platform_ids(yaml_data)

    for platform_id in platform_ids:
        print(platform_id)


if __name__ == "__main__":
    main()
