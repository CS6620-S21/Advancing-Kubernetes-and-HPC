dnf config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo
dnf install https://download.docker.com/linux/centos/8/x86_64/stable/Packages/containerd.io-1.4.3-3.1.el8.x86_64.rpm
dnf install docker-ce
systemctl enable docker
systemctl start docker
