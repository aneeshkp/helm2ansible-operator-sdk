package templating

import (
	"bytes"
	"io/ioutil"
	"text/template"
	"log"
)

type Data struct {
	DataFields []string
}
func RenderMainTaskTemplate(templateFiles []string, dst string) error{
	var tpl = `---
      	- name: Create resources for {{ name }}
	  k8s:
	   state: present
	   definition: "{{lookup}}"
	   loop:
	     {{- range .DataFields }}
              - name: {{.}}
	     {{- end }}`

	funcMap := template.FuncMap{
		"name":   func() string { return "{{name}}" },
		"lookup": func() string { return `{{ lookup('template', item.name) | from_yaml }}` },
	}
		tmpl, err := template.New("").Funcs(funcMap).Parse(tpl)
	if err != nil {
		log.Fatal("Error Parsing template: ", err)
		return err
	}

	var tplOut bytes.Buffer
	if err := tmpl.Execute(&tplOut,Data{DataFields: templateFiles}); err != nil {
		return err
	}
	//templateString:= tmpl.ExecuteTemplate(os.Stdout, "",Data{DataFields: templateFiles} )
	err = ioutil.WriteFile(dst, tplOut.Bytes(), 0644)
	if err != nil {
		log.Fatal("Error Parsing task template: ", err)
	}
	return nil
}