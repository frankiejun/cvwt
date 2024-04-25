#!/bin/bash

if [ $# -lt 1 ]; then
  echo "start.sh config_file_path [ipfile]"
  exit 0
fi

configfile=$1;
ipfile=$2;

echo "configfile:$configfile, ipfile:$ipfile"

ipzipfile="txt.zip"

if [[ -e $ipzipfile ]]; then
  rm -rf $ipzipfile;
#  rm -rf *.csv
fi;

echo "0.读取配置文件"
if [[ ! -e $configfile ]]; then
  echo "找不到$configfile配置文件!"
  exit -1
fi

x_email=$(yq eval ".x_email" $configfile)
hostname=$(yq eval ".hostname" $configfile)
zone_id=$(yq eval ".zone_id" $configfile)
api_key=$(yq eval ".api_key" $configfile)
pause=$(yq eval ".pause" $configfile)
clien=$(yq eval ".clien" $configfile)
multip=$(yq eval ".multip" $configfile)
CFST_URL=$(yq eval ".CFST_URL" $configfile)
CFST_N=$(yq eval ".CFST_N" $configfile)
CFST_T=$(yq eval ".CFST_T" $configfile)
CFST_DN=$(yq eval ".CFST_DN" $configfile)
CFST_TL=$(yq eval ".CFST_TL" $configfile)
CFST_TLL=$(yq eval ".CFST_TLL" $configfile)
CFST_SL=$(yq eval ".CFST_SL" $configfile)
CCFLAG=$(yq eval ".CCFLAG" $configfile)
CCODE=$(yq eval ".CCODE" $configfile)
CF_ADDR=$(yq eval ".CF_ADDR" $configfile)
telegramBotToken=$(yq eval ".telegramBotToken" $configfile)
telegramBotUserId=$(yq eval ".telegramBotUserId" $configfile)

IFS=, read -r -a domains <<< "$hostname";
IFS=, read -r -a countryCodes <<< "$CCODE";

domain_num=${#domains[@]}
countryCode_num=${#countryCodes[@]}

if [ ${#domains[@]} -eq 0 ]; then
	echo "hostname must be set in config file!";
	exit -1;
fi

#检查域名和国家代码是否一一对应
if [ "$CCFLAG" = "true" ]; then
  echo "domain_num:$domain_num, countryCode_num:$countryCode_num"
	if [ $domain_num -ne $countryCode_num ]; then
		echo "The name and country code must correspond one to one!";
    exit -1;
  fi
fi;


GetProxName(){
  c=$1
  if  [ "$c" = "6" ] ; then
  	CLIEN=bypass;
  elif  [ "$c" = "5" ] ; then
  		CLIEN=openclash;
  elif  [ "$c" = "4" ] ; then
  	CLIEN=clash;
  elif  [ "$c" = "3" ] ; then
  		CLIEN=shadowsocksr;
  elif  [ "$c" = "2" ] ; then
  			CLIEN=passwall2;
  			else
  			CLIEN=passwall;
  fi
  echo $CLIEN
}

handle_err() {
  echo "Restore background process."
  proxy=$(GetProxName $clien)
  /etc/init.d/$proxy start
}

trap handle_err EXIT
CLIEN=$(GetProxName $clien)
ps -ef | grep $CLIEN | grep -v "grep" > /dev/null

if [ $? = 1 ]; then
    /etc/init.d/$CLIEN start
fi

if [ -z $ipfile ]; then
  echo "1.Download ip file."
  for i in {1..3}
  do
  	wget  -O $ipzipfile https://zip.baipiao.eu.org
  	
  	if [ $? != 0 ]; then
  	  echo "get ip file failed, try again"
      sleep 1
  	  continue
  	else
      echo "downloaded."
      break
    fi
  done
  
   
  if [ -e $ipzipfile ]; then
    unzip -o $ipzipfile
  else
    echo "Can't download the ip zip file, Check whether the agent software is disabled."
  fi
  
  
  echo "2.Select the ip address of the desired port."
  port=$( yq eval ".CF_ADDR" $configfile)
  if [ -z $port ];then
    port=443
  fi;
  
  for file in $(find . -type f -name "*-[0-1]-$port.txt"); do
      echo "handling: $file"
      cat "$file" >> tmp.txt
  done
  
  if [ -e tmp.txt ]; then
    cat tmp.txt | sort -u > ip.txt
    rm -rf tmp.txt
  fi
fi

echo "Run scripts to test speed and update dns records."
source cf_ddns
