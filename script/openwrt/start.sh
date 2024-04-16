#!/bin/bash

ipzipfile="txt.zip"

if [[ -e $ipzipfile ]]; then
  rm -rf $ipzipfile;
fi;

echo "0.读取配置文件"
if [[ ! -e config.yaml ]]; then
  echo "找不到config.yaml配置文件!"
  exit -1
fi

x_email=$(yq eval ".x_email" config.yaml)
hostname=$(yq eval ".hostname" config.yaml)
zone_id=$(yq eval ".zone_id" config.yaml)
api_key=$(yq eval ".api_key" config.yaml)
pause=$(yq eval ".pause" config.yaml)
clien=$(yq eval ".clien" config.yaml)
CFST_URL=$(yq eval ".CFST_URL" config.yaml)
CFST_N=$(yq eval ".CFST_N" config.yaml)
CFST_T=$(yq eval ".CFST_T" config.yaml)
CFST_DN=$(yq eval ".CFST_DN" config.yaml)
CFST_TL=$(yq eval ".CFST_TL" config.yaml)
CFST_TLL=$(yq eval ".CFST_TLL" config.yaml)
CFST_SL=$(yq eval ".CFST_SL" config.yaml)
CCFLAG=$(yq eval ".CCFLAG" config.yaml)
CCODE=$(yq eval ".CCODE" config.yaml)
CF_ADDR=$(yq eval ".CF_ADDR" config.yaml)
telegramBotToken=$(yq eval ".telegramBotToken" config.yaml)
telegramBotUserId=$(yq eval ".telegramBotUserId" config.yaml)

IFS=, read -r -a domains <<< "$hostname";
IFS=, read -r -a countryCodes <<< "$CCODE";

domain_num=${#domains[@]}
countryCode_num=${#countryCodes[@]}

if [ ${#domains[@]} -eq 0 ]; then
	echo "hostname must be set in config file!";
	exit -1;
fi

#检查域名和国家代码是否一一对应
if [ ! -z $CCFLAG ]; then
  echo "domain_num:$domain_num, countryCode_num:$countryCode_num"
	if [ $domain_num -ne $countryCode_num ]; then
		echo "The name and country code must correspond one to one!";
    exit -1;
  fi
fi;


handle_err() {
  echo "Restore background process."
  if  [ "$clien" = "6" ] ; then
  	CLIEN=bypass;
  elif  [ "$clien" = "5" ] ; then
  		CLIEN=openclash;
  elif  [ "$clien" = "4" ] ; then
  	CLIEN=clash;
  elif  [ "$clien" = "3" ] ; then
  		CLIEN=shadowsocksr;
  elif  [ "$clien" = "2" ] ; then
  			CLIEN=passwall2;
  			else
  			CLIEN=passwall;
  fi
  /etc/init.d/$CLIEN start
}

trap handle_err ERR

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
fi


echo "2.Select the ip address of the desired port."
port=$( yq eval ".CF_ADDR" config.yaml)
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


echo "Run scripts to test speed and update dns records."
source cf_ddns
