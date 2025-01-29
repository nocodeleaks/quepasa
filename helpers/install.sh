#!/bin/bash
MINGOVERSION=1.22
GOVERSIONTOINSTALL=1.22.9

echo 'Installation tested on fresh Ubuntu (20.04|22.04) (ARM64|AMD64)'

echo 'Installing GCC'
apt install gcc -y &>/dev/null

GOVERSION=`go version 2>/dev/null`
if [[ "${GOVERSION}" != *"${MINGOVERSION}"* ]]; then
  
    echo 'Installing GO language'

    # for rpm versions rpm --eval '%{_arch}'
    ARCH=`dpkg --print-architecture`

    echo "Installing for arch ${ARCH}"
    wget "https://go.dev/dl/go${GOVERSIONTOINSTALL}.linux-${ARCH}.tar.gz" -q -O /usr/src/golang.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf /usr/src/golang.tar.gz
    GOROOT=/usr/local/go
    GOPATH=$HOME/go
    PATH=$PATH:$GOROOT/bin
    ln -sf ${GOROOT}/bin/go /usr/sbin/go
    sed -nir '/^export GOROOT=/!p;$a export GOROOT='${GOROOT} ~/.bashrc
    sed -nir '/^export GOPATH=/!p;$a export GOPATH='${GOPATH} ~/.bashrc
    sed -nir '/^export PATH=/!p;$a export PATH='${PATH}:${GOROOT}/bin ~/.bashrc
    GOVERSION=`go version`

fi
echo "Installed ${GOVERSION}"

echo 'Updating Quepasa link'
ln -sf /opt/quepasa-source/src /opt/quepasa 

echo 'Ensuring RSyslog'
apt install rsyslog -y &>/dev/null

echo 'Updating logging'
ln -sf /opt/quepasa-source/helpers/syslog.conf /etc/rsyslog.d/10-quepasa.conf

echo 'Updating log rotate'
ln -sf /opt/quepasa-source/helpers/quepasa.logrotate.d /etc/logrotate.d/quepasa

/bin/mkdir -p /var/log/quepasa
/bin/chmod 755 /var/log/quepasa
/bin/chown syslog:adm /var/log/quepasa

echo 'Restarting services'
systemctl restart rsyslog

echo 'Updating systemd service'
ln -sf /opt/quepasa-source/helpers/quepasa.service /etc/systemd/system/quepasa.service
systemctl daemon-reload

adduser --disabled-password --gecos "" --home /opt/quepasa quepasa
chown -R quepasa /opt/quepasa-source

cp /opt/quepasa-source/helpers/.env /opt/quepasa/.env

systemctl enable quepasa.service
systemctl start quepasa

# Hint: Setup Quepasa user
echo 'Setup Quepasa user >>>  http://<your-ip>:31000/setup'

exit 0
