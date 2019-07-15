// Package jnlp provides data structures and ability for parsing JNLP files.
package jnlp

import (
	"encoding/xml"
	"io/ioutil"
	"net/url"

	launcher_utils "github.com/rocketsoftware/open-web-launch/launcher/utils"
	"github.com/rocketsoftware/open-web-launch/utils/log"
)

// JNLP is a main xml element for a jnlp file
type JNLP struct {
	CodeBase          string       `xml:"codebase,attr"`
	Spec              string       `xml:"spec,attr,omitempty"`
	Href              string       `xml:"href,attr,omitempty"`
	Version           string       `xml:"version,attr,omitempty"`
	Information       *Information `xml:"information"`
	Resources         []*Resources `xml:"resources"`
	AppDescription    *AppDesc     `xml:"application-desc"`
	AppletDescription *AppletDesc  `xml:"applet-desc"`
}

// Resources that are needed for an application
type Resources struct {
	OS         string       `xml:"os,attr,omitempty"`     // operating system
	Arch       string       `xml:"arch,attr,omitempty"`   // architecture
	Locale     string       `xml:"locale,attr,omitempty"` // locales
	JARs       []*JAR       `xml:"jar,omitempty"`
	Properties []Property   `xml:"property,omitempty"`
	Extensions []*Extension `xml:"extension,omitempty"`
	J2SE       *J2SE        `xml:"j2se,omitempty"`
	Java       *J2SE        `xml:"java,omitempty"` // synonym for j2se
	NativeLibs []*NativeLib `xml:"nativelib,omitempty"`
}

// JAR file that is part of the application's classpath
type JAR struct {
	Href     string `xml:"href,attr"`               // URL of the jar file
	Version  string `xml:"version,attr,omitempty"`  // The requested version of the jar file. Requires using the version-based download protocol
	Main     bool   `xml:"main,attr,omitempty"`     // Indicates if this JAR file contains the class containing the main method of the RIA
	Download string `xml:"download,attr,omitempty"` // Indicates that this JAR file can be downloaded lazily, or when needed
	Size     int64  `xml:"size,attr,omitempty"`     // The downloadable size of the JAR file in bytes
	Part     bool   `xml:"part,attr,omitempty"`     // Can be used to group resources together so that they are downloaded at the same time
}

// NativeLib is a JAR file that contains native libraries in its root directory.
type NativeLib struct {
	Href     string `xml:"href,attr"`               // URL of the jar file
	Version  string `xml:"version,attr,omitempty"`  // The requested version of the jar file. Requires using the version-based download protocol
	Main     bool   `xml:"main,attr,omitempty"`     // Indicates if this JAR file contains the class containing the main method of the RIA
	Download string `xml:"download,attr,omitempty"` // Indicates that this JAR file can be downloaded lazily, or when needed
	Size     int64  `xml:"size,attr,omitempty"`     // The downloadable size of the JAR file in bytes
	Part     bool   `xml:"part,attr,omitempty"`     // Can be used to group resources together so that they are downloaded at the same time
}

// Extension is a pointer to an additional component-desc or installer-desc to be used with this RIA
type Extension struct {
	Href    string `xml:"href,attr"`              // The URL to the additional extension JNLP file
	Version string `xml:"version,attr,omitempty"` // The version of the additional extension JNLP file.
	Name    string `xml:"name,attr,omitempty"`    // The name of the additional extension JNLP file
	URL     string `xml:"url,attr,omitempty"`     // Resolved Href reference - only for developer usage
}

// J2SE represents versions of Java software to run the RIA with
type J2SE struct {
	Version         string `xml:"version,attr"`                     // Ordered list of version ranges to use.
	Href            string `xml:"href,attr,omitempty"`              // The URL denoting the supplier of this version of Java software, and from where it can be downloaded.
	JavaVMArgs      string `xml:"java-vm-args,attr,omitempty"`      // An additional set of standard and non-standard virtual machine arguments that the RIA would prefer the JNLP client use when launching the JRE software.
	InitialHeapSize string `xml:"initial-heap-size,attr,omitempty"` // The initial size of the Java heap.
	MaxHeapSize     string `xml:"max-heap-size,attr,omitempty"`     // The maximum size of the Java heap.
}

// Information contains other elements that describe the RIA and its source
type Information struct {
	OS             string          `xml:"os,attr,omitempty"`         // The operating system for which this information element should be considered
	Arch           string          `xml:"arch,attr,omitempty"`       // The architecture for which this information element should be considered
	Platform       string          `xml:"platform,attr,omitempty"`   // The platform for which this information element should be considered
	Locale         string          `xml:"locale,attr,omitempty"`     // The locale for which this information element should be considered.
	Title          string          `xml:"title"`                     // The title of the RIA
	Vendor         string          `xml:"vendor"`                    // The provider of the RIA
	Homepage       *Homepage       `xml:"homepage,omitempty"`        // The homepage of the RIA.
	Descriptions   []*Description  `xml:"description,omitempty"`     // Short statements describing the RIA
	Icons          []*Icon         `xml:"icon,omitempty"`            // An icon that can be used to identify the RIA to the user
	Shortcut       *Shortcut       `xml:"shortcut,omitempty"`        // Can be used to indicate the RIA's preferences for desktop integration.
	Desktop        *xml.Name       `xml:"desktop"`                   // Can be used to indicate the RIA's preference for putting a shortcut on the user's desktop
	Menu           *Menu           `xml:"menu,omitempty"`            // Can be used to indicate the RIA's preference for putting a menu item in the user's start menus
	OfflineAllowed *OfflineAllowed `xml:"offline-allowed,omitempty"` // Indicates that this application can operate when the client system is disconnected from the network.
	Version        string          `xml:"version,omitempty"`         // Application version
}

// Homepage of the RIA
type Homepage struct {
	Href string `xml:"href,attr"` // A URL pointing to where more information about this RIA can be found.
}

// Description is a short statement describing the RIA.
type Description struct {
	Kind string `xml:"kind,attr,omitempty"` // Indicator as to the type of description. Legal values are one-line, short, and tooltip.
	Text string `xml:",chardata"`           // Description text
}

// Icon that can be used to identify the RIA to the user
type Icon struct {
	Href       string `xml:"href,attr"`                 // A URL pointing to the icon file. Can be in one of the following formats: gif, jpg, png, ico.
	Kind       string `xml:"kind,attr,omitempty"`       // Indicates the suggested use of the icon, can be: default, selected, disabled, rollover, splash, or shortcut.
	Downloaded bool   `xml:"downloaded,attr,omitempty"` // For internal usage: indicates that the icon successfully downloaded
}

// Shortcut can be used to indicate the RIA's preference for putting a shortcut on the user's desktop
type Shortcut struct {
	Online  bool      `xml:"online,attr,omitempty"` // Can be used to describe the RIA's preference for creating a shortcut to run online or offline.
	Desktop *xml.Name `xml:"desktop"`               // Can be used to indicate the RIA's preference for putting a shortcut on the user's desktop
	Menu    *Menu     `xml:"menu,omitempty"`        // Can be used to indicate the RIA's preference for putting a menu item in the user's start menus
}

// Menu can be used to indicate the RIA's preference for putting a menu item in the user's start menus.
type Menu struct {
	SubMenu string `xml:"submenu,attr,omitempty"` // Can be used to indicate the RIA's preference for where to place the menu item.
}

type OfflineAllowed struct {
}

// Property defines a system property that will be available through
// the System.getProperty and System.getProperties methods
type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// AppDesc denotes this is the JNLP file for an application
type AppDesc struct {
	MainClass string   `xml:"main-class,attr"` // The name of the class containing the public static void main(String[]) method of the application
	Arguments []string `xml:"argument"`        // Each argument contains (in order) an additional argument to be passed to the main method
}

// AppletDesc denotes this is the JNLP file for an applet
type AppletDesc struct {
	MainClass    string     `xml:"main-class,attr"`             // The name of the main applet class
	DocumentBase string     `xml:"documentbase,attr,omitempty"` // The document base for the applet as a URL
	Name         string     `xml:"name,attr"`                   // Name of the applet
	Width        int        `xml:"width,attr"`                  // The width of the applet in pixels
	Height       int        `xml:"height,attr"`                 // The height of the applet in pixels
	Params       []Property `xml:"param"`                       // A set of parameters that can be passed to the applet
}

// Decode decodes JNLP data
func Decode(data []byte) (*JNLP, error) {
	var jnlp JNLP
	if err := xml.Unmarshal(data, &jnlp); err != nil {
		return nil, err
	}
	return &jnlp, nil
}

// DecodeFile decodes JNLP file
func DecodeFile(filename string) (*JNLP, error) {
	var jnlp JNLP
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := xml.Unmarshal(data, &jnlp); err != nil {
		return nil, err
	}
	return &jnlp, nil
}

// RelevantForCurrentPlatform returns true if resources are actual for current platform
func (resources *Resources) RelevantForCurrentPlatform() bool {
	return launcher_utils.AreResourcesRelevantForCurrentPlatform(resources.OS, resources.Arch)
}

func (jnlp *JNLP) findRelevantResources() []*Resources {
	var relevantResources []*Resources
	for _, resources := range jnlp.Resources {
		if resources.RelevantForCurrentPlatform() {
			log.Printf("resources for os='%s' arch='%s' are relevant on current platform", resources.OS, resources.Arch)
			relevantResources = append(relevantResources, resources)
		} else {
			log.Printf("resources for os='%s' arch='%s' are not relevant on current platform", resources.OS, resources.Arch)
		}
	}
	return relevantResources
}

func (jnlp *JNLP) getJars() ([]string, error) {
	var urls []string
	codebaseURL, err := launcher_utils.ParseCodebaseURL(jnlp.CodeBase)
	if err != nil {
		return nil, err
	}
	relevantResources := jnlp.findRelevantResources()
	for _, resources := range relevantResources {
		for _, jar := range resources.JARs {
			url, err := url.Parse(jar.Href)
			if err != nil {
				continue
			}
			abs := codebaseURL.ResolveReference(url)
			urls = append(urls, abs.String())
		}
	}
	return urls, nil
}

func (jnlp *JNLP) getNativeLibs() ([]string, error) {
	var urls []string
	codebaseURL, err := launcher_utils.ParseCodebaseURL(jnlp.CodeBase)
	if err != nil {
		return nil, err
	}
	relevantResources := jnlp.findRelevantResources()
	for _, resources := range relevantResources {
		for _, jar := range resources.NativeLibs {
			url, err := url.Parse(jar.Href)
			if err != nil {
				continue
			}
			abs := codebaseURL.ResolveReference(url)
			urls = append(urls, abs.String())
		}
	}
	return urls, nil
}

// Title returns title of JNLP application
// or empty string if Information tag not found
func (jnlp *JNLP) Title() string {
	if jnlp.Information != nil {
		return jnlp.Information.Title
	}
	return ""
}

func (resources *Resources) getJ2SE() *J2SE {
	if resources.J2SE != nil {
		return resources.J2SE
	}
	if resources.Java != nil {
		return resources.Java
	}
	return nil
}
