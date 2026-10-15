Add-Type -AssemblyName UIAutomationClient, UIAutomationTypes
Add-Type 'using System;using System.Runtime.InteropServices;public class U{[DllImport("user32.dll")]public static extern IntPtr SendMessage(IntPtr h,int m,IntPtr w,IntPtr l);}'
$dir = Split-Path $args[0]
function Mine { Get-Process uxplay -EA SilentlyContinue | ? Path -like "$dir*" }
$p = Start-Process $args[0] -PassThru; Start-Sleep 4
$root = [System.Windows.Automation.AutomationElement]::FromHandle($p.MainWindowHandle)
function All { $root.FindAll([System.Windows.Automation.TreeScope]::Descendants, [System.Windows.Automation.Condition]::TrueCondition) }
function Btn { All | ? { $_.Current.Name -cin "Start","Stop","Mulai","Berhenti" -and $_.Current.NativeWindowHandle -ne 0 } | Select -First 1 }
# custom-drawn button: no BM_CLICK, send a left click at (10,10)
function Click { $h = [IntPtr](Btn).Current.NativeWindowHandle; [void][U]::SendMessage($h, 0x0201, [IntPtr]1, [IntPtr]0xA000A); [void][U]::SendMessage($h, 0x0202, [IntPtr]::Zero, [IntPtr]0xA000A) }
"checks: " + (All | % { $_.Current.Name } | ? { $_ -match "(of|dari) \d (ready|siap)" } | Select -First 1)
Click; Start-Sleep 5
"after Start: button=" + (Btn).Current.Name + " uxplay=" + (Mine | Measure).Count
Click; Start-Sleep 2
"after Stop:  uxplay=" + (Mine | Measure).Count
Click; Start-Sleep 3
"restart:     uxplay=" + (Mine | Measure).Count
[void]$p.CloseMainWindow(); Start-Sleep 2
"after Close: uxplay=" + (Mine | Measure).Count + " app alive=" + (-not $p.HasExited)
Mine | Stop-Process -Force
