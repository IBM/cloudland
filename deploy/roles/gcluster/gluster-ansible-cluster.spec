%global docdir %{_datadir}/doc/gluster.cluster
%global rolesdir %{_sysconfdir}/ansible/roles/gluster.cluster

Name:      gluster-ansible-cluster
Version:   0.1
Release:   1%{?dist}
Summary:   Ansible roles for GlusterFS volume management

URL:       https://github.com/gluster/gluster-ansible-cluster
Source0:   %{url}/archive/v%{version}.tar.gz#/%{name}-%{version}.tar.gz
License:   GPLv3
BuildArch: noarch

Requires:  ansible >= 2.6

%description
Collection of Ansible roles for the creating and managing GlusterFS volumes.

%prep
%setup -q -n %{name}-%{version}

%build

%install
mkdir -p %{buildroot}/%{docdir}
install -p -m 644 README.md %{buildroot}/%{docdir}
cp -r examples %{buildroot}/%{docdir}/

mkdir -p %{buildroot}/%{rolesdir}
cp -dpr defaults handlers meta roles tasks tests LICENSE vars \
   %{buildroot}/%{rolesdir}

%files
%doc %{docdir}
%rolesdir

%license LICENSE

%changelog
* Fri Aug 31 2018 Sachidananda Urs <sac@redhat.com> 0.1
- Initial release, volume creation and set options
