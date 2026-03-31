#!/bin/zsh

tf_files=${PROJECT_ROOT}/internal/provider

for res_file in "${tf_files}"/*_resource.go
do
  if [[ "$res_file" =~ provider\/([^\/]+)_resource.go ]]; then
      domain=${BASH_REMATCH[1]}
      echo "Domain name: $domain"
      mkdir -p examples/resources/${domain}
      touch examples/resources/${domain}/resource.tf
  else
      echo "Invalid file ${res_file}"
  fi
done

for ds_file in "${tf_files}"/*_data_source.go
do
  if [[ "$ds_file" =~ provider\/([^\/]+)_data_source.go ]]; then
      domain=${BASH_REMATCH[1]}
      echo "Domain name: $domain"
      mkdir -p examples/data-sources/${domain}
      touch examples/data-sources/${domain}/data-source.tf
  else
      echo "Invalid file ${ds_file}"
  fi
done


printf "#----------arena resources ---------#\n" > examples/tf/res.tf

for eg_file in examples/resources/**/*.tf
do
  cat ${eg_file} >> examples/tf/res.tf
done