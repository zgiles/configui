# configui
Router autoconfiguration tool for Mikrotik routers, especially for mesh networks

[![Build Status](https://travis-ci.com/zgiles/configui.svg?token=G9G1zGy4KwRkpSJ7QvjQ&branch=master)](https://travis-ci.com/zgiles/configui)

---

# Use
Run the program, allow it to autodetect or enter an address for a fresh router. Select if you want to update firmware and/or config. Click Start  
No external programs are required, all functionality is internal to the program ( no ssh, etc )   

Notes:
- Only autodiscovers mikrotik routers
- Currently only works on a reset-to-default routers
- You do still need to get the config from configgen, and the firmware file from the website. Place them in your downloads folder next to the program for it to discover them.  
- OSX only looks in your User's Downloads folder for now


## Mac
There's an "app" version, but it isnt "signed" so when you try and run it, it will warn you.  
If you're on OSX <10.15, you can Option-Click and press OK, Above you need to: Go to System Preferences > Security&Privacy > Allow configui. Then try to run the app again.  
Alternately, there is a "unpackaged" version you can launch from the command line and will load open a GUI.  

( Command line launch: Open the compressed file you downloaded. A nwe file will appear. Open Terminal. Run "cd ~/Downloads" ( no quotes ). Then run "./configui-osx-v1.0" )


## Linux
Download the released version, run, no problem.
Download the tar ball, there's a binary inside, works fine.

## Windows
Dont have a windows computer, but it should compile and run on windows
The windows version is built by CI. It builds and runs, but I'm not sure if it functions properly.. it might not have. Let me know.

## License 
Copyright (c) 2020 Zachary Giles Licensed under the MIT/Expat License
See LICENSE

