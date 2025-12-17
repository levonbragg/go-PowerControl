# Parse .env file and generate batch commands
# This script reads a .env file and outputs batch file SET commands

$envPath = "..\\.env"

if (Test-Path $envPath) {
    $envContent = Get-Content $envPath
    
    foreach ($line in $envContent) {
        # Skip empty lines and comments
        if ($line -match '^\s*$' -or $line -match '^\s*#') {
            continue
        }
        
        # Parse KEY=VALUE
        if ($line -match '^([^=]+)=(.*)$') {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            
            # Remove surrounding quotes if present
            $value = $value.Trim('"')
            
            # Output batch SET command
            Write-Output "set `"$key=$value`""
        }
    }
}
