Add-Type -AssemblyName UIAutomationClient, UIAutomationTypes
Add-Type 'using System;using System.Runtime.InteropServices;public class U{[DllImport("user32.dll")]public static extern IntPtr SendMessage(IntPtr h,int m,IntPtr w,IntPtr l);}'
$dir = Split-Path $args[0]
function Mine { Get-Process uxplay -EA SilentlyContinue | ? Path -like "$dir*" }
$p = Start-Process $args[0] -PassThru; Start-Sleep 4
$root = [System.Windows.Automation.AutomationElement]::FromHandle($p.MainWindowHandle)
function All { $root.FindAll([System.Windows.Automation.TreeScope]::Descendants, [System.Windows.Automation.Condition]::TrueCondition) }
function Btn { All | ? { $_.Current.Name -in "Start","Stop" -and $_.Current.NativeWindowHandle -ne 0 } | Select -First 1 }
function Click { [void][U]::SendMessage([IntPtr](Btn).Current.NativeWindowHandle, 0x00F5, [IntPtr]::Zero, [IntPtr]::Zero) }
"doctor OK lines: " + ((All | % { $_.Current.Name }) -join "`n" | Select-String -AllMatches "\[OK\]").Matches.Count
Click; Start-Sleep 5
"after Start: button=" + (Btn).Current.Name + " uxplay=" + (Mine | Measure).Count
Click; Start-Sleep 2
"after Stop:  uxplay=" + (Mine | Measure).Count
Click; Start-Sleep 3
"restart:     uxplay=" + (Mine | Measure).Count
[void]$p.CloseMainWindow(); Start-Sleep 2
"after Close: uxplay=" + (Mine | Measure).Count + " app alive=" + (-not $p.HasExited)
Mine | Stop-Process -Force
