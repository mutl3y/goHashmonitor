package profitSwitch

var Debug bool

func main() {

}

/*
	Function check-Profit-Stats{
			Param ([Parameter( Position=0, Mandatory, ValueFromPipeline )]$coins,
			       [Parameter( Position=1, Mandatory, ValueFromPipeline )][int]$hr, [switch]$silent
			)

			$statsURL="https://minecryptonight.net/api/rewards?hr=$hr&limit=0"
			$uridata=$null
			$path="$ScriptDir\profit.json"
			$data=@{}
			if( $coins ){$supportedCoins=$coins.ToUpper()}
			$script:bestcoins=@{}

			function get-stats{
				try{
					$uridata=Invoke-WebRequest -UseBasicParsing -Uri $statsURL -TimeoutSec 60
					$uridata | Set-Content -Path $path
				}
				catch{
					log-write -logstring "Issue updating profit stats, Using last set" -for red -notification 2
					Check-Network
				}
			}

			# Refresh stats file
			if( ! (Test-Path -path $path ) ){
				get-stats
			}
			else{
				$test=Get-Item $path |
				      Where-Object{$_.LastWriteTime-lt(Get-Date ).AddSeconds( - $proftStatRefreshTime )}
				if( $test ){
					get-stats
					write-host "Profit stats refreshed from https://minecryptonight.net/api/rewards "
				}
			}

			#Read from profit.json
			$rawdata=(Get-Content -RAW -Path $path | Out-String | ConvertFrom-Json )
			$script:pools=@{} # Clean current hashtable
			#Add each coin to an ordered list
			foreach( $coin in $rawdata.rewards ){
				if( ($coin.ticker_symbol ) | Where-Object ({$_-in$supportedCoins} ) ){
					switch( $coin.algorithm ){
						{$_-in"cryptonightv2"}     {
							$script:pools.Add( $coin.ticker_symbol,
							                   [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-v2_factor' ) )*$cryptonightv2_factor ),
							                                                    10 ) )
						}
						{$_-in"cryptonight-fast"}     {
							$script:pools.Add( $coin.ticker_symbol,
							                   [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-fast_factor' ) )*$cryptonightfast_factor ),
							                                                    10 ) )
						}
						{$_-in"cryptonight-heavy", "cryptonight-saber", 'cryptonight-haven', 'cryptonight-webchain'}     {
							$script:pools.Add( $coin.ticker_symbol,
							                   [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-heavy_factor' ) )*$cryptonightheavy_factor ),
							                                                    10 ) )
						}
						{$_-in"cryptonight-lite", "cryptonight-lite-v1"}     {
							$script:pools.Add( $coin.ticker_symbol,
							                   [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-lite_factor' ) )*$cryptonightlite_factor ),
							                                                    10 ) )
						}
						Default {
							$script:pools.Add( $coin.ticker_symbol, [ decimal ][System.Math]::Round( $coin.reward_24h.btc, 10 ) )
						}
					}

				}
				else{
					switch( $coin.algorithm ){
						{$_-in"cryptonight-v2"}     {
							$script:bestcoins.Add( $coin.ticker_symbol,
							                       [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-v2_factor' ) )*$cryptonightv2_factor ),
							                                                        10 ) )
						}
						{$_-in"cryptonight-fast"}     {
							$script:bestcoins.Add( $coin.ticker_symbol,
							                       [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-fast_factor' ) )*$cryptonightfast_factor ),
							                                                        10 ) )
						}
						{$_-in"cryptonight-heavy", "cryptonight-saber", 'cryptonight-haven', 'cryptonight-webchain'}     {
							$script:bestcoins.Add( $coin.ticker_symbol,
							                       [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-heavy_factor' ) )*$cryptonightheavy_factor ),
							                                                        10 ) )
						}
						{$_-in"cryptonight-lite", "cryptonight-lite-v1"}     {
							$script:bestcoins.Add( $coin.ticker_symbol,
							                       [ decimal ][System.Math]::Round( (($coin.reward_24h.btc/($rawdata.'cryptonight-lite_factor' ) )*$cryptonightlite_factor ),
							                                                        10 ) )
						}
						Default {
							$script:bestcoins.Add( $coin.ticker_symbol,
							                       [ decimal ][System.Math]::Round( $coin.reward_24h.btc, 10 ) )
						}
					}
				}
			}

			#Check our pools
			if( $script:pools ){
				$bestcoin=($script:bestcoins.ValueSort() ).GetEnumerator() | Select-Object -first 1
				$ourcoin=($script:pools.ValueSort() ).GetEnumerator() | Select-Object -first 1
				$profitLoss=$bestcoin.value-$ourcoin.value
				if( ! ($silent ) ){
					log-write -logstring "Coin Selected, $( $ourcoin.Name ) " -fore green -notification 1
					log-write -logstring "Possible earnings per day from stats with a min hashrate of $hr H/s for our enabled coins`n" -fore yellow -attachment ($script:pools.ValueSort() ).ToDisplayString() -notification 2
					log-write -logstring "Difference in Daily earnings:$profitLoss BTC Per day Mining $( $bestcoin.Name )" -fore yellow -notification 2
				}

				# Export coin to mine to script
				$script:coinToMine=$ourcoin.Name
			}
			else{
				log-write -logstring   "No compatable entries found in $ScriptDir\pools.txt" -fore red -notification 1
			}
		}



*/
