package kubeExplorer

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type commandType struct {
	Name string
	f    func(g *gocui.Gui, v *gocui.View) error
}

type keyEventType struct {
	Viewname string
	Key      interface{}
	mod      gocui.Modifier
}

type keyBindingType struct {
	KeyEvent keyEventType
	Command  commandType
}

func newScaleCommand(name string, replicas int) commandType {
	var scaleCommand = commandType{Name: name, f: func(g *gocui.Gui, v *gocui.View) error {
		res := currentResource()
		if res.Name == "deployments" || res.Name == "replicationcontrollers" || res.Name == "replicasets" || res.Name == "daemonsets" || res.Name == "statefulsets" {
			tmp := resourceItemsList.widget.title
			resourceItemsList.widget.title = fmt.Sprintf("Scaling to %v replicas...", replicas)
			scaleResource(replicas)
			backend.resetCache()
			g.Update(func(gui *gocui.Gui) error {
				newResource()
				resourceItemsList.widget.title = tmp
				return nil
			})
		}
		return nil
	}}
	return scaleCommand
}

func newExecCommand(name, cmd string, containerNr int) commandType {
	var execCommand = commandType{Name: name, f: func(g *gocui.Gui, v *gocui.View) error {
		res := currentResource()
		if res.Name != "pods" {
			return nil
		}
		ns := currentNamespace()
		rname := currentResourceItemName()

		pod := resourceItemsList.widget.items[resourceItemsList.widget.selectedItem]
		containers := val(pod, []interface{}{"spec", "containers"}, "")
		containerNames := toStringArray(containers, "name")
		if containerNr < len(containerNames) {
			in, out, err := backend.execIntoPod(ns, rname, cmd, containerNames[containerNr], func() {
				g.Cursor = false
				setState(browseState)
			})
			if err != nil {
				showError("Can't exec into pod", err)
				return nil
			}
			setState(execPodState)
			execWidget.title = fmt.Sprintf("exec container '%s' in pod '%s'", containerNames[containerNr], rname)
			execWidget.open(g, in, out)
		}
		return nil
	}}
	return execCommand
}

func newPortForwardCommand(name string, useSamePort bool) commandType {

	var portForwardCommand = commandType{Name: name, f: func(g *gocui.Gui, v *gocui.View) error {
		res := currentResource()
		if res.Name != "pods" {
			return nil
		}
		podName := currentResourceItemName()
		ns := currentNamespace()
		if portforwardProxies[ns+"/"+podName] != nil {
			err := removePortforwardProxyofPod(podName)
			if err != nil {
				showError("Can't remove port-forward proxy", err)
				return nil
			}
			return nil
		}
		pod := resourceItemsList.widget.items[resourceItemsList.widget.selectedItem]

		ports := ports(pod)
		for _, cp := range ports {
			var localPort int
			if useSamePort {
				localPort = cp.port
			} else {
				localPort = currentPortforwardPort
				currentPortforwardPort++
			}
			err := createPortforwardProxy(podName, portMapping{localPort, cp})
			if err != nil {
				showError("Can't create port-forward proxy", err)
				return nil
			}
		}
		return nil
	}}
	return portForwardCommand
}

var portForwardSamePortCommand = newPortForwardCommand("toggle port forward (same port)", true)
var portForwardCommand = newPortForwardCommand("toggle port forward", false)

var quitCommand = commandType{Name: "Quit", f: func(g *gocui.Gui, v *gocui.View) error {
	removeAllPortforwardProxies()
	return gocui.ErrQuit
}}

var nextResourceCommand = commandType{Name: "Next resource", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceMenu.widget.nextSelectedItem()
	newResource()
	return nil
}}
var previousResourceCommand = commandType{Name: "Previous resource", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceMenu.widget.previousSelectedItem()
	newResource()
	return nil
}}

var reloadCommand = commandType{Name: "Reload", f: func(g *gocui.Gui, v *gocui.View) error {
	backend.resetCache()
	newResource()
	return nil
}}

var deleteCommand = commandType{Name: "Delete resource", f: func(g *gocui.Gui, v *gocui.View) error {
	tmp := resourceItemsList.widget.title
	resourceItemsList.widget.title = "Deleting pod ..."
	deleteResource()
	backend.resetCache()
	g.Update(func(gui *gocui.Gui) error {
		newResource()
		resourceItemsList.widget.title = tmp
		return nil
	})
	return nil
}}

var nameSortCommand = commandType{Name: "Sort by name", f: func(g *gocui.Gui, v *gocui.View) error {
	nameSorting()
	newResource()
	return nil
}}

var ageSortCommand = commandType{Name: "Sort by age", f: func(g *gocui.Gui, v *gocui.View) error {
	ageSorting()
	newResource()
	return nil
}}

var scaleUpCommand = newScaleCommand("Scale up", 1)
var scaleDownCommand = newScaleCommand("Scale down", -1)

var execBashCommand0 = newExecCommand("Exec first container bash", "bash", 0)
var execBashCommand1 = newExecCommand("Exec second container bash", "bash", 1)
var execBashCommand2 = newExecCommand("Exec third container bash", "bash", 2)
var execShellCommand0 = newExecCommand("Exec first container sh", "sh", 0)
var execShellCommand1 = newExecCommand("Exec second container sh", "sh", 1)
var execShellCommand2 = newExecCommand("Exec third container sh", "sh", 2)

var nextResourceItemDetailPartCommand = commandType{Name: "Next resource", f: func(g *gocui.Gui, v *gocui.View) error {
	resourcesItemDetailsMenu.widget.nextSelectedItem()
	setResourceItemDetailsPart()
	return nil
}}

var previousResourceItemDetailPartCommand = commandType{Name: "Next resource", f: func(g *gocui.Gui, v *gocui.View) error {
	resourcesItemDetailsMenu.widget.previousSelectedItem()
	setResourceItemDetailsPart()
	return nil
}}

var nextLineCommand = commandType{Name: "Next resource item", f: func(g *gocui.Gui, v *gocui.View) error {
	if resourceItemDetailsWidget.visible {
		return nil
	}
	resourceItemsList.widget.nextSelectedItem()
	return nil
}}
var previousLineCommand = commandType{Name: "Previous resource item", f: func(g *gocui.Gui, v *gocui.View) error {
	if resourceItemDetailsWidget.visible {
		return nil
	}
	resourceItemsList.widget.previousSelectedItem()
	return nil
}}

var nextPageCommand = commandType{Name: "Next resource item page ", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemsList.widget.nextPage()
	return nil
}}

var previousPageCommand = commandType{Name: "Previous resource item page ", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemsList.widget.previousPage()
	return nil
}}

var toggleResourceItemDetailsCommand = commandType{Name: "Toggle resource item details ", f: func(g *gocui.Gui, v *gocui.View) error {
	toggleDetailBrowseState()
	return nil
}}

var findNextCommand = commandType{Name: "find next on resource item details ", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemDetailsWidget.findNext()
	return nil
}}

var findPreviousCommand = commandType{Name: "find previous on resource item details ", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemDetailsWidget.findPrevious()
	return nil
}}

var homeCommand = commandType{Name: "TextArea home", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemDetailsWidget.scrollUp(1<<63 - 1)
	return nil
}}

var endCommand = commandType{Name: "TextArea end", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemDetailsWidget.scrollDown((1<<63 - 1) / 2)
	return nil
}}

var pageUpCommand = commandType{Name: "TextArea page up", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemDetailsWidget.scrollUp(resourceItemDetailsWidget.h)
	return nil
}}

var pageDownCommand = commandType{Name: "TextArea page down", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemDetailsWidget.scrollDown(resourceItemDetailsWidget.h)
	return nil
}}

var scrollUpCommand = commandType{Name: "TextArea scroll up", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemDetailsWidget.scrollUp(1)
	return nil
}}

var scrollDownCommand = commandType{Name: "TextArea scroll down", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemDetailsWidget.scrollDown(1)
	return nil
}}

var scrollUpHelpCommand = commandType{Name: "Help scroll up", f: func(g *gocui.Gui, v *gocui.View) error {
	helpWidget.scrollUp(1)
	return nil
}}

var scrollDownHelpCommand = commandType{Name: "Help scroll down", f: func(g *gocui.Gui, v *gocui.View) error {
	helpWidget.scrollDown(1)
	return nil
}}

var scrollRightCommand = commandType{Name: "TextArea scroll right", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemDetailsWidget.scrollRight()
	return nil
}}
var scrollLeftCommand = commandType{Name: "TextArea scroll left", f: func(g *gocui.Gui, v *gocui.View) error {
	resourceItemDetailsWidget.scrollLeft()
	return nil
}}

var gotoSelectContextStateCommand = commandType{Name: "Select context", f: func(g *gocui.Gui, v *gocui.View) error {
	setState(selectContextState)
	return nil
}}

var selectContextCommand = commandType{Name: "Select context", f: func(g *gocui.Gui, v *gocui.View) error {
	clusterList.widget.title = "Loading resources of cluster (%d/%d)...    "
	g.Update(func(gui *gocui.Gui) error {
		newContext()
		clusterList.widget.title = "[C]luster"
		return nil
	})
	setState(browseState)
	return nil
}}

var nextContextCommand = commandType{Name: "Next context", f: func(g *gocui.Gui, v *gocui.View) error {
	clusterList.widget.nextSelectedItem()
	return nil
}}
var previousContextCommand = commandType{Name: "Previous context", f: func(g *gocui.Gui, v *gocui.View) error {
	clusterList.widget.previousSelectedItem()
	return nil
}}

var gotoSelectNamespaceStateCommand = commandType{Name: "Select namespace", f: func(g *gocui.Gui, v *gocui.View) error {
	setState(selectNsState)
	return nil
}}

var selectNamespaceCommand = commandType{Name: "Select namespace", f: func(g *gocui.Gui, v *gocui.View) error {
	namespaceList.widget.title = "Loading resources of namespace (%d/%d)...    "
	g.Update(func(gui *gocui.Gui) error {
		findResourceCategoryWithResources(1)
		namespaceList.widget.title = "[N]amespace"
		return nil
	})
	setState(browseState)
	return nil
}}

var nextNamespaceCommand = commandType{Name: "Next namespace", f: func(g *gocui.Gui, v *gocui.View) error {
	namespaceList.widget.nextSelectedItem()
	return nil
}}

var previousNamespaceCommand = commandType{Name: "Previous namespace", f: func(g *gocui.Gui, v *gocui.View) error {
	namespaceList.widget.previousSelectedItem()
	return nil
}}

var nextResourceCategoryCommand = commandType{Name: "Next resource category", f: func(g *gocui.Gui, v *gocui.View) error {
	nextResourceCategory(1)
	findResourceCategoryWithResources(1)
	return nil
}}

var previousResourceCategoryCommand = commandType{Name: "Next resource category", f: func(g *gocui.Gui, v *gocui.View) error {
	nextResourceCategory(-1)
	findResourceCategoryWithResources(-1)
	return nil
}}

var showHelpCommand = commandType{Name: "Show help", f: func(g *gocui.Gui, v *gocui.View) error {
	setState(helpState)
	return nil
}}

// var execCommand = commandType{Name: "Exec in to pod", f: func(g *gocui.Gui, v *gocui.View) error {
// 	res := currentResource()
// 	if res.Name != "pods" {
// 		return nil
// 	}
// 	ns := currentNamespace()
// 	rname := currentResourceItemName()
// 	in, out, err := backend.execIntoPod(ns, rname, "bin/sh", func() {
// 		g.Cursor = false
// 		setState(browseState)
// 	})
// 	if err != nil {
// 		showError("Can't exec into pod",err)
// 		return nil
// 	}
// 	setState(execPodState)
// 	execWidget.title = fmt.Sprintf("exec in %s",rname)
// 	execWidget.open(g, in, out)
// 	return nil
// }}

var quitWidgetCommand = commandType{Name: "quit help", f: func(g *gocui.Gui, v *gocui.View) error {
	setState(browseState)
	return nil
}}

var keyBindings = []keyBindingType{}

func bindKey(g *gocui.Gui, keyBind keyEventType, command commandType) {
	if err := g.SetKeybinding(keyBind.Viewname, keyBind.Key, keyBind.mod, command.f); err != nil {
		errorlog.Panicln(err)
	}
	kb := keyBindingType{keyBind, command}
	keyBindings = append(keyBindings, kb)
}
