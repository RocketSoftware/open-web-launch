# Open Web Launch

## Introduction

For years, Java Web Start has been used as deployment vehicle for Java Desktop applications. As of Java 9, however, the feature has been deprecated – and in Java 11, Java Web Start has been removed completely.

Oracle has pitched several other deployment scenarios, but many existing projects, products and components have trouble to make that change in time, or at all for that matter. Open Web Launch has been created to fill the gap this change of functionality leaves behind.

**Note:** 

Open Web Launch can work with any Java version of any provider \(Oracle, OpenJDK or IBM\), with JREs and JDKs. Note that Open Web Launch will not address any Java-compatibility issues for the Java applications it serves – this is the responsibility of their manufacturer. The prime goal for Open Web Launch is to run any application as configured in its JNLP file against a Java version which may officially no longer support Java Web Start.

## Usage Scenarios

### JNLP files

This scenario makes sure double clicking a JNLP file in the explorer opens it, downloading required resources, and starting the application as instructed.

**Note:** 

This scenario needs the Open Web Launch \(OWL\) application to be installed.

### JNLP URLs

This scenario takes care of intercepting JNLP file URLs that are clicked on in a browser and redirects their handling to the Open Web Launch application, downloading required resources, and starting the application as instructed.

**Note:** 

This scenario needs the OWL browser extension and application to be installed.

### JNLP protocol

This scenario redirects every URI starting with `jnlp:` or `jnlps:` and redirects their handling to the Open Web Launch application, downloading required resources, and starting the application as instructed.

**Note:** 

This scenario needs the OWL application to be installed.

## Installation

There are two ways to install OWL on your system – either through a setup \(executable\) or through a browser extension.

### Prerequisite

An appropriate Java version needs to be installed on the system for OWL to work. Certificates required by a Java application need to be imported.

### Setup

The setup allows to specify some configuration options. These can be modified post-installation by running **Modify** from the Control Panel or by choosing **Configure Open Web Launch** from the Start menu.

#### User install

Run the setup for the current user only.

-   Point at the Java you want to use for all Web Start applications

-   Select whether you want to make OWL the default for opening JNLP files

-   Select whether you want to register the JNLP and JNLPS protocol for Open Web Start

-   Select whether you want to show the Java console when opening JNLP files

#### Admin install

Run the setup so that all users on the system have access to it.

-   Point at the Java you want to use for all Web Start applications

-   Select whether you want to make OWL the default for opening JNLP files

-   Select whether you want to register the JNLP and JNLPS protocol for Open Web Start

-   Select whether you want to show the Java console when opening JNLP files

**Note:** 

Administrative privileges are required for this.

#### Silent install

There is an option to run the setup silently, which uses the defaults:

`setup /s`

#### Uninstall

OWL can be uninstalled from the Control Panel or from a shortcut in the Start menu. 

### Browser extension

#### Chrome

The extension for Chrome is available in the Chrome Web Store from [https://chrome.google.com/webstore/detail/open-web-launch/pmmlhpkdpbddohdbnjinopbkmlcnjnhc](https://chrome.google.com/webstore/detail/open-web-launch/pmmlhpkdpbddohdbnjinopbkmlcnjnhc).

#### **Firefox**

The add-on for Firefox is available on the Mozilla site from [https://addons.mozilla.org/en-US/firefox/addon/open-web-launch/](https://addons.mozilla.org/en-US/firefox/addon/open-web-launch/).

##Command Line Operations

Open Web Launch has the following command line options:

**Default**

This is the command line executed when double-clicking a JNLP file.

`openweblaunch.exe <jnlp reference>`

**-Uninstall**

This command allows to uninstall a specific Java Web Start application.

`openweblaunch.exe -uninstall <jnlp reference>`

**-JavaDir**

This command allows to pass a specific Java that should be used for starting a Java Web Start application.

`openweblaunch.exe -javadir <java folder> <jnlp reference>`

## Appendix

### Frequently Asked Questions

#### What operating systems does Open Web Launch support?

Open Web Launch is available for Windows, macOS and Linux.

#### Are there 32 and 64-bit versions available?

Both versions are installed by default (`openweblaunch32.exe` and `openweblaunch64.exe`).
Based on the JVM selected, the setup will make the 32 or 64-bit version of Open Web Launch the default. 

#### What happens if a JNLP file on the host changes?

Open Web Launch will check for changes between remote and local JNLP files and refresh where needed.

#### How does Open Web Launch determine the Java it should use

This is the order by which Open Web Launch determines what Java executable it will use to run a Java Web Start application:

**Command line**

When `-JavaDir` is specified on the command line, this is the version of Java which will be used.

**JAVA_HOME**

Open Web Launch will use the `JAVA_HOME` environment variable to locate the version of Java it should use if this was selected during setup.

**Registry**

Open Web Launch will use a specific version of Java if this was indicated during setup.

**Path**

If none of the other options result in a Java version that it can use, Open Web Launch will try to locate Java on the `PATH`.


### Supported keywords

|Element|   |Attribute|Values / Description|
|-------|---|---------|--------------------|
|**information**| | | |
| |icon| | |
| |shortcut| | |
|**title**| | | |
| |vendor| | |
| |homepage| | |
| |description| | |
|**application-desc**| | | |
|**resources**| | | |
| | |os|windows, darwin, linux|
| | |arch|amd64, x86|
| |jar| | |
| | |href| |
| |nativelib| | |
| | |href| |
| | |name| |
| |extension| | |
| | |href| |
| | |name| |
| | |version| |

### Links

#### JNLP definition

[https://docs.oracle.com/javase/tutorial/deployment/deploymentInDepth/jnlp.html](https://docs.oracle.com/javase/tutorial/deployment/deploymentInDepth/jnlp.html)

#### JNLP examples

[https://docs.oracle.com/javase/tutorial/uiswing/examples/misc/index.html](https://docs.oracle.com/javase/tutorial/uiswing/examples/misc/index.html)

#### Chrome Extension (beta)

[https://chrome.google.com/webstore/detail/open-web-launch/pmmlhpkdpbddohdbnjinopbkmlcnjnhc](https://chrome.google.com/webstore/detail/open-web-launch/pmmlhpkdpbddohdbnjinopbkmlcnjnhc)

#### Firefox Add-on (beta)

[https://addons.mozilla.org/en-US/firefox/addon/open-web-launch/](https://addons.mozilla.org/en-US/firefox/addon/open-web-launch/)
