package algo

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
)

// Rhel9_6Installer represent the installer implementation for rhel 9.6 os distribution
type Rhel9_6Installer struct {
	install   string
	uninstall string
}

// NewRhel9_6Installer will return new Rhel9_6Installer instance
func NewRhel9_6Installer(ctx context.Context, arch, bundleAddrs string) (*Rhel9_6Installer, error) {
	parseFn := func(script string) (string, error) {
		parser, err := template.New("parser").Parse(script)
		if err != nil {
			return "", fmt.Errorf("unable to parse install script")
		}
		var tpl bytes.Buffer
		if err = parser.Execute(&tpl, map[string]string{
			"BundleAddrs":          bundleAddrs,
			"Arch":                 arch,
			"ImgpkgVersion":        ImgpkgVersion,
			"BundleDownloadPath":   "{{.BundleDownloadPath}}",
			"ContainerdConfigToml": GetConfigTomlEchoString(),
		}); err != nil {
			return "", fmt.Errorf("unable to apply install parsed template to the data object")
		}
		return tpl.String(), nil
	}

	install, err := parseFn(DoRhel9_6K8s)
	if err != nil {
		return nil, err
	}
	uninstall, err := parseFn(UndoRhel9_6K8s)
	if err != nil {
		return nil, err
	}
	return &Rhel9_6Installer{
		install:   install,
		uninstall: uninstall,
	}, nil
}

// Install will return k8s install script
func (s *Rhel9_6Installer) Install() string {
	return s.install
}

// Uninstall will return k8s uninstall script
func (s *Rhel9_6Installer) Uninstall() string {
	return s.uninstall
}

// contains the installation and uninstallation steps for the supported os and k8s
var (
	DoRhel9_6K8s = `
set -euox pipefail

BUNDLE_DOWNLOAD_PATH={{.BundleDownloadPath}}
BUNDLE_ADDR={{.BundleAddrs}}
IMGPKG_VERSION={{.ImgpkgVersion}}
ARCH={{.Arch}}
BUNDLE_PATH=$BUNDLE_DOWNLOAD_PATH/$BUNDLE_ADDR
CONTAINERD_CONFIG_TOML='{{.ContainerdConfigToml}}'

# Install imgpkg if not available
if ! command -v imgpkg >>/dev/null; then
	echo "installing imgpkg"	
	
	if command -v wget >>/dev/null; then
		dl_bin="wget -nv -O-"
	elif command -v curl >>/dev/null; then
		dl_bin="curl -s -L"
	else
		echo "installing curl"
		apt-get install -y curl
		dl_bin="curl -s -L"
	fi
	
	$dl_bin github.com/vmware-tanzu/carvel-imgpkg/releases/download/$IMGPKG_VERSION/imgpkg-linux-$ARCH > /tmp/imgpkg
	mv /tmp/imgpkg /usr/bin/imgpkg
	chmod +x /usr/bin/imgpkg
fi

echo "downloading bundle"
mkdir -p $BUNDLE_PATH
imgpkg pull -i $BUNDLE_ADDR -o $BUNDLE_PATH


## disable swap
swapoff -a && sed -ri '/\sswap\s/s/^#?/#/' /etc/fstab

## Disable firewall
if systemctl is-active firewalld &>/dev/null; then
    systemctl disable --now firewalld
fi

## load kernal modules
modprobe overlay && modprobe br_netfilter

## adding os configuration
tar -C / -xvf "$BUNDLE_PATH/conf.tar" && sysctl --system 

## installing RPM packages
for pkg in cri-tools kubernetes-cni kubectl kubelet kubeadm; do
    dnf install -y "$BUNDLE_PATH/$pkg.rpm"
done

## prevent updates to kubelet/kubectl/kubeadm
dnf mark install cri-tools kubernetes-cni kubectl kubelet kubeadm

## intalling containerd
tar -C / -xvf "$BUNDLE_PATH/containerd.tar"

## setup containerd config file
if [ ! -e /etc/containerd/config.toml ]; then
    mkdir -p /etc/containerd
    echo "$CONTAINERD_CONFIG_TOML" > /etc/containerd/config.toml
    chmod 755 /etc/containerd && chmod 644 /etc/containerd/config.toml
fi


## starting containerd service
systemctl daemon-reload && systemctl enable containerd && systemctl start containerd`

	UndoRhel9_6K8s = `
set -euox pipefail

BUNDLE_DOWNLOAD_PATH={{.BundleDownloadPath}}
BUNDLE_ADDR={{.BundleAddrs}}
BUNDLE_PATH=$BUNDLE_DOWNLOAD_PATH/$BUNDLE_ADDR

## disabling containerd service
systemctl stop containerd && systemctl disable containerd && systemctl daemon-reload

## removing containerd configurations and cni plugins
rm -rf /opt/cni/ && rm -rf /opt/containerd/ &&  tar tf "$BUNDLE_PATH/containerd.tar" | xargs -n 1 echo '/' | sed 's/ //g'  | grep -e '[^/]$' | xargs rm -f

## removing RPM packages (Kubernetes components)
for pkg in kubeadm kubelet kubectl kubernetes-cni cri-tools; do
    if dnf list installed "$pkg" >/dev/null 2>&1; then
        dnf remove -y "$pkg"
    fi
done

## removing os configuration
tar tf "$BUNDLE_PATH/conf.tar" | xargs -n 1 echo '/' | sed 's/ //g' | grep -e "[^/]$" | xargs rm -f

## unload kernal modules
modprobe -rq overlay && modprobe -r br_netfilter

## enable firewall if firewalld is available
if systemctl is-enabled firewalld &>/dev/null; then
	systemctl enable firewalld
	systemctl start firewalld
fi

## enable swap
swapon -a && sed -ri '/\sswap\s/s/^#?//' /etc/fstab

rm -rf $BUNDLE_PATH`
)
