package main

import (
  "fmt"
  "os"
  "os/exec"
  "strings"
  "io/ioutil"
  tea "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/lipgloss"
  "golang.org/x/exp/slices"
)

var (
  titleStyle = lipgloss.NewStyle().
    Background(lipgloss.Color("#505059")).
    Foreground(lipgloss.Color("#ffffff"))
  disclaimerStyle= lipgloss.NewStyle().
    Background(lipgloss.Color("#505059")).
    Foreground(lipgloss.Color("#ffffff")).
    PaddingLeft(2)
  bodyStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#404040"))
  warningStyle = lipgloss.NewStyle().
    Background(lipgloss.Color("#ffffff")).
    Foreground(lipgloss.Color("#a03000"))
  
  quitStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#a0a0a0")).
    PaddingLeft(3).
    PaddingRight(3)
  commandStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#0050c0"))

  borderStyle = lipgloss.NewStyle().
    BorderStyle(lipgloss.NormalBorder()).
    BorderForeground(lipgloss.Color("#000000")).
    PaddingLeft(3).
    PaddingRight(3)

)


type model struct {
  altscreen bool
  command string
  commandPresent bool
  commandType string
  err error 
  disclaimerShow bool
  testMenu bool
}

type commandFinishedMsg struct { err error }

func (m model) Init() tea.Cmd {
  return nil
}

func getHDD(m model) (tea.Model, tea.Cmd){
  hdd_fusion_check := exec.Command("bash", "-c", "diskutil info /dev/disk2")
  // dev/disk0 being the default drive macOS is installed on
  hdd_disk0 := exec.Command("bash", "-c", "diskutil info /dev/disk0 | grep \"Disk Size\" | awk '{print $3, $4, $5, $6}'")
  // dev/disk2 being the default assignment a fusion drive gets. (Comprising of a disk0 SSD and a disk1 HDD for example)
  hdd_disk2 := exec.Command("bash", "-c", "diskutil info /dev/disk2 | grep \"Disk Size\" | awk '{print $3, $4, $5, $6}'")

  // Boring Go error handling
  fusion_info_bytes, err := hdd_fusion_check.Output()
  if err != nil {
    fmt.Println("Error! ", err)
  }
  hdd_info, err := hdd_disk0.Output()
  if err != nil {
    fmt.Println("Error!", err)
  }
  hdd_info2, err := hdd_disk2.Output()
  if err != nil {
    fmt.Println("Error!", err)
  }

  // If "Fusion Drive" is found in the output of hdd_fusion_check, then print the disk size of /dev/disk2 (the default allocation
  //  of a fusion drive), otherwise print the disk size of /dev/disk0
  fusion_info_string := string(fusion_info_bytes[:])
  if strings.Contains(fusion_info_string, "Fusion Drive"){
    m.command = bodyStyle.Render("\nDEVICE IS USING FUSION DRIVE\n/dev/disk2: ") + commandStyle.Render(string(hdd_info2[:]))
  } else {
    m.command = bodyStyle.Render("   /dev/disk0: ") + commandStyle.Render(string(hdd_info[:]))
  }

  // **IGNORE**
  // m.command = bodyStyle.Render("\n/dev/disk0: ") + commandStyle.Render(string(hdd_info[:])) + bodyStyle.Render("\n/dev/disk1: ")+ commandStyle.Render(string(hdd_info2[:])) + bodyStyle.Render("\n/dev/disk2: ") + commandStyle.Render(string(hdd_info3[:]))
  m.commandType = "HDD"
  return m, nil
}


func getGPU(m model) (tea.Model, tea.Cmd){
  c:= exec.Command("bash", "-c", "ioreg -rc IOPCIDevice | grep \"model\" | sed -n '1 p'")
  gpu_info, err := c.Output() 

  // Boring Go error handling
  if err != nil {
    fmt.Println("Error Getting GPU:", err)
    return m, nil
  }

  m.command = string(gpu_info[:]) 
  m.commandType = "GPU"
  return m, nil

}

func getRAM(m model) (tea.Model, tea.Cmd) {
  c := exec.Command("bash", "-c", "sysctl hw.memsize | awk '{print $2/1024/1024/1024 \"GB\"}'")
  ram_info, err := c.Output()

  // Boring Go error handling
  if err != nil {
    fmt.Println("Error Getting RAM:", err)
    return m, nil
  }

  m.command = (string(ram_info[:]))
  m.commandType = "RAM"
  return m, nil

}

func getCPU(m model) (tea.Model, tea.Cmd) {
  c := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
  cpu_info, err := c.Output()

  // Boring Go error handling
  if err != nil {
    fmt.Println("Error Getting CPU: ", err)
    return m, nil
  }

  m.command = string(cpu_info[:])
  m.commandType = "CPU"
  return m, nil
}

func installOS(m model) (tea.Model, tea.Cmd){
  // m.command = warningStyle.Render("This action must be allowed to complete in it's entirety before anything else is done. Are you sure? (Y/N)")

  rootFolder, err := ioutil.ReadDir("/")
  if err != nil {
    fmt.Println("Error reading '/' directory ", err)
    return m, nil
  } 

  var rootFolderItems []string

  for _, f := range rootFolder {
    rootFolderItems = append(rootFolderItems, f.Name())
  }

  if slices.Contains(rootFolderItems, "Install macOS Catalina.app"){
    m.commandType = "OLDOS"
    return m, nil
  }

  
  c := exec.Command("bash", "-c", "./install_os.sh")

  install_os_info, err := c.Output()

  if err != nil {
    fmt.Println("Error Installing OS:", err)

    return m, nil
  }

  m.command = string(install_os_info[:])
  m.commandType = "OS"

  return m, nil
}

func getWifi(m model) (tea.Model, tea.Cmd) {
  c := exec.Command("/usr/libexec/airportd", "en1", "alloc", "--ssid", "Geoff", "--security", "wpa2", "--password", "digital1")
  wifi_info, err := c.Output()

  // Boring Go error handling
  if err != nil {
    fmt.Println("Error Connecting to WiFi: ", err)
    return m, nil
  }

  m.command = string(wifi_info[:])
  m.commandType = "WIFI"
  return m, nil
}

func pingTest(m model) (tea.Model, tea.Cmd){
  c := exec.Command("ping", "-c", "1", "www.google.com")
  ping_test, err := c.Output()

  // Boring Go error handling
  if err != nil {
    fmt.Println("Error pinging https://google.com : ", err)
    return m, nil
  }

  m.command = string(ping_test[:])
  m.commandType = "PING"

  return m, nil
}

func formatDrive(m model, fs string) (tea.Model, tea.Cmd) {
  if fs == "APFS" { 
    c := exec.Command("bash", "-c", "diskutil erasedisk APFS \"Macintosh HD\" /dev/disk0")
    m.command = "APFS"
    err := c.Run()
    if err != nil {
      fmt.Println("Error formating drive to APFS: ", err)
      return m, nil
    }
  } else if fs == "JHFS+" {
    c := exec.Command("bash", "-c", "diskutil erasedisk JHFS+ \"Macintosh HD\" /dev/disk0")
    m.command = "JHFS+"
    err := c.Run()
    if err != nil {
      fmt.Println("Error formatting drive to JHFS+: ", err)
      return m, nil
    }
  } else {
    c := exec.Command("diskutil", "resetfusion")
    m.command = "FUSION"
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr
    err := c.Run()
    if err != nil {
      fmt.Println("Error formatting fusion drive: ", err)
      return m, nil
    }
  }
  m.commandType = "FORMAT"
  return m, nil
}

func testMenu_HDDWriteTest(m model) (tea.Model, tea.Cmd) {
  // c := exec.Command("bash", "-c", "dd if=/dev/zero bs=2048k of=tstFile count=1024 2>&1 | grep sec | awk '{print $1 / 1024 / 1024 / $5, \"MB/sec\"}')")
  c := exec.Command("dd", "if=/dev/zero", "bs=2048k", "of=tstFile2", "count=1024", "|", "grep", "sec", "|", "awk", "'{print $1/ 1024/ 1024/ $5, \"MB/sec\"}'")
  // err := c.Run()
  writeTest_info, err := c.Output()
  fmt.Println("HDD Write Test engaged")
  if err != nil {
    fmt.Println("Error executing write test: ", err)
    return m, nil
  }
  fmt.Println("HDD Write Test completed")
  m.commandType = "TEST_WRITE"
  m.command = string(writeTest_info[:])
  return m, nil
}

func testMenu_HDDReadTest(m model) (tea.Model, tea.Cmd) {
  return m, nil
}


func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type){
  case tea.KeyMsg:
    switch msg.String(){
    case "q":
      return m, tea.Quit
    case "o":
      m.commandPresent = true
      m.disclaimerShow = false
      return(installOS(m))

    case "h":
      m.commandPresent = true
      m.disclaimerShow = false
      if m.testMenu {
        m.testMenu = !m.testMenu
        return testMenu_HDDWriteTest(m)
      }
      return getHDD(m)
    case "c":
      m.commandPresent = true
      m.disclaimerShow = false
      return getCPU(m)
    case "r":
      m.commandPresent = true
      m.disclaimerShow = false
      if m.testMenu {
        m.testMenu = !m.testMenu
        return m, nil
      }
      return getRAM(m)
    case "g":
      m.commandPresent = true
      m.disclaimerShow = false
      return getGPU(m)
    case "w":
      m.commandPresent = true
      m.disclaimerShow = false
      if m.testMenu {
        m.testMenu = !m.testMenu
        return m, nil
      }
      return getWifi(m)
    case "p":
      m.commandPresent = true
      m.disclaimerShow = false
      return pingTest(m)
    case "1":
      m.commandPresent = true
      m.disclaimerShow = false
      return formatDrive(m, "APFS")
    case "2":
      m.commandPresent = true
      m.disclaimerShow = false
      return formatDrive(m, "JHFS+")
    case "3":
      m.commandPresent = true
      m.disclaimerShow = false
      return formatDrive(m, "Fusion")
    case "t":
      m.commandPresent = true
      m.disclaimerShow = false
      m.commandType = "TEST"
      m.testMenu = true
      return m, nil
    case "b":
      m.commandPresent = false
      m.disclaimerShow = false
      return m, nil
    case "a":
      m.disclaimerShow = false
      m.altscreen = !m.altscreen
      cmd := tea.EnterAltScreen
      if !m.altscreen{
        cmd = tea.ExitAltScreen
      }
      return m, cmd
    case "x":
      m.disclaimerShow = false
      return m, nil
    }
  case commandFinishedMsg:
    if msg.err != nil {
      m.err = msg.err
      return m, tea.Quit
    }
  }
  return m, nil
}

func (m model) View() string {
  var renderString, titleString string
  var displayOptions bool
  displayOptions = true
  renderString = ""
  titleString = ""
  if m.disclaimerShow {
    titleString += disclaimerStyle.Render("This is Niall's shitty hardware detection tool. No support given") + "\n"
  }
  if !m.commandPresent {
    renderString += titleStyle.Render("Please enter a command")

    renderString += "\n" + bodyStyle.Render("'c' for ") + commandStyle.Render("cpu")
    renderString += "\n" + bodyStyle.Render("'r' for ") + commandStyle.Render("ram") 
    renderString += "\n" + bodyStyle.Render("'g' for ") + commandStyle.Render("gpu") 
    renderString += "\n" + bodyStyle.Render("'h' for ") + commandStyle.Render("hdd")
    renderString += "\n" + bodyStyle.Render("'w' for ") + commandStyle.Render("wifi ") + warningStyle.Render("BROKEN")
    renderString += "\n" + bodyStyle.Render("'o' for ") + commandStyle.Render("os install ") + warningStyle.Render("NOT IMPLEMENTED")
    renderString += "\n" + bodyStyle.Render("'p' for ") + commandStyle.Render("ping test")
    renderString += "\n" + bodyStyle.Render("'1' for ") + commandStyle.Render("APFS Format")
    renderString += "\n" + bodyStyle.Render("'2' for ") + commandStyle.Render("JHFS+ Format")
    renderString += "\n" + bodyStyle.Render("'3' for ") + commandStyle.Render("Fusion Drive Format ") + warningStyle.Render("NOT IMPLEMENTED")
    renderString += "\n" + bodyStyle.Render("'t' for ") + commandStyle.Render("Test Menu")
    renderString += "\n" + quitStyle.Render("'q' to quit")
  } else {
    renderString = "\n"
    switch m.commandType {
    case "GPU":
      renderString += bodyStyle.Render("GPU is: ") + commandStyle.Render(m.command) + "\n" + warningStyle.Render("TODO: FORMAT THIS BETTER")

    case "RAM":
      renderString += bodyStyle.Render("RAM is: ") + commandStyle.Render(m.command)
    case "CPU":
      renderString += bodyStyle.Render("CPU is: ") + commandStyle.Render(m.command)
    case "HDD":
      renderString += bodyStyle.Render("Disk Size: ") + commandStyle.Render(m.command)
    case "OS":
      renderString += m.command + "\n\n"
      displayOptions = false
    case "PING":
      renderString += bodyStyle.Render("Ping results: \n") + commandStyle.Render(m.command)
    case "FORMAT":
      renderString += bodyStyle.Render("Formatted drive: ") + commandStyle.Render(m.command)
    case "WIFI":
      renderString += bodyStyle.Render("If wifi card was detected, you should now be connected!")
      displayOptions = false
    case "TEST":
      renderString += bodyStyle.Render("Yo sorry B but this hasn't been implemented yet. Look out for these tests in the future")
      renderString += "\n" + bodyStyle.Render("'???' for ") + commandStyle.Render("Harddrive read test")

      renderString += "\n" + bodyStyle.Render("'???' for ") + commandStyle.Render("Harddrive write test")
      renderString += "\n" + bodyStyle.Render("'???' for ") + commandStyle.Render("CPU Stress test")
      renderString += "\n" + bodyStyle.Render("'???' for ") + commandStyle.Render("RAM test")
    case "TEST_WRITE":
      renderString += "\n" + bodyStyle.Render("Write Test:") + commandStyle.Render(m.command)
    case "OLDOS":
      renderString += "\n" + bodyStyle.Render("The OS you're booting off of is too old for this command to work.") + "\n" + bodyStyle.Render("Please use ") + commandStyle.Render("Big Sur") + bodyStyle.Render(" or later to use this feature")

    default:
      renderString += bodyStyle.Render("Fucking uhhhhhh") + commandStyle.Render(" Idk B")
    }
    m.commandPresent = !m.commandPresent
    if displayOptions {
      renderString += "\n" + quitStyle.Render("[c]pu | [g]pu | [r]am | [h]dd | [b]ack to menu | [q]uit")
    }
  }
  /*
  if m.command.Contains("GB") {
    m.commandPresent = !m.commandPresent
  } else if m.command.Contains("Intel") {
    renderString = "\n"
    m.commandPresent = !m.command
  }
  */

  if m.err != nil {
    return "Error: " + m.err.Error() + "\n"
  }
  outputString := titleString + borderStyle.Render(renderString)
  titleString, renderString = "", ""
  return outputString

}

func main(){
  m := model{commandPresent: false, disclaimerShow: true, testMenu: false}

  if err:= tea.NewProgram(m).Start(); err != nil {
    fmt.Println("Error! ", err)
    os.Exit(1)
  }
}
