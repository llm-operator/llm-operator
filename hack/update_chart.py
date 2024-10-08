# This script updates the dependencies in the Helm chart file.
#
# Usage:
#   python update_chart.py <filename>


import subprocess
import sys

def get_latest_version(repo):
    cmds = [
        'git',
        'ls-remote',
        '--tags',
        '--refs',
        '--sort=v:refname',
        'https://github.com/llmariner/%s.git' % repo
    ]
    output = subprocess.check_output(cmds).decode('utf-8')
    tags = output.split('\n')
    tags.reverse()
    latest_tagline = ''
    for tag in tags:
        if tag:
            latest_tagline = tag
            break
    latest_tag = latest_tagline.split('\t')[1].split('/')[-1]
    return latest_tag.strip('v')


def update_chart(filename):
    repos = [
        'api-usage',
        'cluster-manager',
        'file-manager',
        'inference-manager',
        'job-manager',
        'model-manager',
        'user-manager',
        'rbac-manager',
        'session-manager',
        'vector-store-manager',
    ]
    vers = {}
    for repo in repos:
        ver = get_latest_version(repo)
        vers[repo] = ver

    deps = {
        'api-usage-server': vers['api-usage'],
        'cluster-manager-server': vers['cluster-manager'],
        'dex-server': vers['rbac-manager'],
        'file-manager-server': vers['file-manager'],
        'inference-manager-engine': vers['inference-manager'],
        'inference-manager-server': vers['inference-manager'],
        'job-manager-dispatcher': vers['job-manager'],
        'job-manager-server': vers['job-manager'],
        'model-manager-loader': vers['model-manager'],
        'model-manager-server': vers['model-manager'],
        'rbac-server': vers['rbac-manager'],
        'session-manager-agent': vers['session-manager'],
        'session-manager-server': vers['session-manager'],
        'user-manager-server': vers['user-manager'],
        'vector-store-manager-server': vers['vector-store-manager'],
    }

    workers = {
        'inference-manager-engine',
        'job-manager-dispatcher',
        'model-manager-loader',
        'session-manager-agent'
    }

    chart = """apiVersion: v2
name: llmariner
description: A Helm chart for LLMariner
type: application
version: 0.1.0
appVersion: 0.1.0
dependencies:
"""
    for dep, ver in deps.items():
        component_type = "worker" if dep in workers else "control-plane"

        chart += """- name: %(dep)s
  version: %(ver)s
  repository: "oci://public.ecr.aws/cloudnatix/llmariner-charts"
  tags:
  - %(component_type)s
""" % {'dep': dep, 'ver': ver, 'component_type': component_type}

    # Write the chart to the file
    with open(filename, 'w') as f:
        f.write(chart)


def main():
    # Get a path name from the commandline arguments
    if len(sys.argv) < 2:
        print("Usage: python %s <filename>" % sys.argv[0])
        sys.exit(1)
    filename = sys.argv[1]
    update_chart(filename)

if __name__ == "__main__":
    main()
