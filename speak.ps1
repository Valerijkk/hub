param(
    [string]$text = "Valera brat!",
    [string]$voice = "Irina"
)

Add-Type -AssemblyName System.Speech
$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer
$synth.Rate = 1

$installed = $synth.GetInstalledVoices()
foreach ($v in $installed) {
    if ($v.VoiceInfo.Name -like "*$voice*") {
        $synth.SelectVoice($v.VoiceInfo.Name)
        break
    }
}

$synth.Speak($text)
