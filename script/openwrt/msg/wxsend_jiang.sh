#!/bin/bash

text=''

while IFS= read -r line; do
  text+="$line"
done <"./informlog"

#for test.
#configfile="./config.yaml"
#sendKey=$(yq eval ".sendKey" $configfile)

title="msg_from_cvwt"
URL="https://sctapi.ftqq.com/$sendKey.send?"

if [[ -z ${sendKey} ]]; then
  echo "未配置微信推送的sendKey, 关注【方糖】公众号获取sendKey"
else
  res=$(timeout 20s curl -s -X POST $URL -d title=${title} -d desp="${text}")
  if [ $? == 124 ]; then
    echo "发送消息超时"
    exit 1
  fi

  #res:{"code":0,"message":"","data":{"pushid":"164999333","readkey":"SCTGLrDh4Zadf4k","error":"SUCCESS","errno":0}}
  err=$(echo "$res" | jq -r ".data.error")
  if [ "$err" == "SUCCESS" ]; then
    echo "微信推送成功"
  else
    echo "微信推送失败, error:$err"
  fi
fi
