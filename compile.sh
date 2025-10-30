export TARGET=usbarmory
make nonsecure_os_go
make trusted_applet_go
make trusted_os
#don't forget to unplug and replug the USB Armory
sudo $HOME/go/bin/armory-boot-usb -i bin/trusted_os_usbarmory.imx
