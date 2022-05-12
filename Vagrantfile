Vagrant.configure("2") do |config|
  config.vm.provider "virtualbox" do |v|
    v.cpus = 2
    v.memory = 2048
  end
  config.vm.box = "ubuntu/jammy64"
  config.vm.synced_folder ".", "/home/vagrant/debugging-with-delve"
  config.vm.network "private_network", type: "dhcp"
  config.vm.provision "shell", path: "provision.sh"

  if Vagrant.has_plugin?("vagrant-vbguest")
    config.vbguest.auto_update = false  
  end
end