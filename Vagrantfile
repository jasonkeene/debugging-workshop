Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/bionic64"
  config.vm.synced_folder ".", "/debugging-workshop", SharedFoldersEnableSymlinksCreate: true
  config.vm.provision "shell", path: "vagrant-provision.sh"
  config.vm.provider "virtualbox" do |v|
    v.memory = 4096
    v.cpus = 2
  end
end
