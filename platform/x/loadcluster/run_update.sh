#!/bin/bash

export ANSIBLE_HOST_KEY_CHECKING=False
export AWS_ACCESS_KEY_ID=AKIAPVZV75IYX5LTLVDA
export AWS_SECRET_ACCESS_KEY=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
export AWS_DEFAULT_REGION=cn-north-1
export AWS_REGION=cn-north-1
ansible-playbook -i aws_hosts.ini locust.yml --private-key ~/.ssh/timesking-awscn.pem --tags "botx" --skip-tags "waithosts"
