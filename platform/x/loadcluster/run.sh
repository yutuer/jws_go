#!/bin/bash

export ANSIBLE_HOST_KEY_CHECKING=False
export AWS_ACCESS_KEY_ID=AKIAOUGT5JMO45QCVMIA
export AWS_SECRET_ACCESS_KEY=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
export AWS_DEFAULT_REGION=cn-north-1
export AWS_REGION=cn-north-1
ansible-playbook -i aws_hosts.ini locust.yml --private-key ~/.ssh/timesking-awscn.pem
#ansible-playbook -i aws_hosts.ini locust.yml --private-key ~/.ssh/timesking-awscn.pem --tags "botx" --skip-tags "waithosts"
