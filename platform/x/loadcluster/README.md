# Swarm for Locust.io

fork from https://github.com/gleicon/locust-swarm.git
a93aa63f5b035b8ee71996cfa09d7802d0b787a0

修改：添加RedHat相关系统的Yum库支持

Amazon Linux AMI 2015.03 (HVM), SSD Volume Type - ami-f239abcb
这个AMI的版本的Python状况：手动检测后发现
- Python2.7
- pip
- virtualenv

locustio on MacOSX
CFLAGS='-std=c99' pip install gevent==1.0.1

locust本身是一个基于gevent单进程的实现。所以在t2.micro这些单核的机器上启动单个进程本身是能够比较合适的。
如果需要利用机器的多核，则必须使用master/slave模式，然后根据机器的vCPU数量启动对应数量的slave


Ansible AWS Provisioning for Locust.io Distributed load testing. There will be one Master VM and a number of slave VMs

## AWS
	- Depends on Ansible and boto
		$ sudo pip install ansible
		$ sudo pip install boto
	- Ships with ec2.py and ec2.ini from ansible. you probably want to check [this doc and update](http://docs.ansible.com/intro_dynamic_inventory.html)
	- Export credentials
		$ export AWS_ACCESS_KEY_ID=<aws access key id>
		$ export AWS_SECRET_ACCESS_KEY=<aws secret access key>
	- Disable host key checking
		$ export ANSIBLE_HOST_KEY_CHECKING=False

	- At AWS, take note of your VPC id, subnet id, IP block assigned to this subnet and the key name you will be using.
	- Each region may have distinct AMI ids for Ubuntu 14.04 64 bits (trusty)
	- Configure group_vars/all with your data
	- slave_count is the number of slave VMs
	- test_host is the target host you will be testing (base host, not full URL)
		ssh_key_name: aws_devel
		aws_region: us-east-1
		ami_id: ami-9eaa1cf6
		instance_type: t2.micro
		vpc_id: vpc-ffffffff
		subnet_id: subnet-ffffffff
		slave_count: 3
		cidr_ip: 10.0.0.0/16
    		test_host: "https://google.com"

	$ ansible-playbook -i aws_hosts.ini locust.yml --private-key ~/.ssh/aws_devel.pem
	- aws_devel.pem is your public key, the same referred in the file above
	- aws_hosts.ini file containing
		[local]
		localhost

	- use EC2 panel to remove servers or $ ansible-playbook -i ./ec2.py remove_servers.yml

## Testing agent

The testing agent is agent.py, copied into templates/ dir of locust role. If you use *run.sh* it will do it for you

## Target host

The target host is configured at locust.yml variable test_host

## Master

Access the master VM using http://ip:8089
