package main

import (
	"encoding/json"
	"github.com/ayushbpl10/protoc-gen-rights/rights"
	"github.com/golang/protobuf/proto"
	"github.com/lyft/protoc-gen-star"
	"github.com/lyft/protoc-gen-star/lang/go"
	"regexp"
	"strings"
)

type rightsGen struct {
	pgs.ModuleBase
	pgsgo.Context
}

func (*rightsGen) Name() string {
	return "zap"
}

func (m *rightsGen) InitContext(c pgs.BuildContext) {
	m.ModuleBase.InitContext(c)
	m.Context = pgsgo.InitContext(c.Parameters())
}

func (m *rightsGen) Execute(targets map[string]pgs.File, packages map[string]pgs.Package) []pgs.Artifact {

	modulePath := "github.com/ayushbpl10/protoc-gen-rights/example/"

	for _, f := range targets {


		name := m.Context.OutputPath(f).SetExt(".rights.go").String()
		fm := fileModel{PackageName: m.Context.PackageName(f).String(), }
		for _,im := range f.Imports() {
			fm.Imports = append(fm.Imports, im.Descriptor().Options.GetGoPackage())
		}

		fm.Imports = append(fm.Imports, modulePath+f.Descriptor().Options.GetGoPackage())


		for _,srv := range f.Services() {

			service := serviceModel{}
			service.ServiceName = srv.Name().String()
			service.PackageName = m.Context.PackageName(f).String()

			for _, rpc := range srv.Methods() {

				opt := rpc.Descriptor().GetOptions()
				option, err := proto.GetExtension(opt, zappb.E_Validator)
				if err != nil {
					panic(err)
				}
				byteData, err := json.Marshal(option)
				if err != nil {
					panic(err)
				}
				rights := zappb.MyRights{}
				err = json.Unmarshal(byteData, &rights)
				if err != nil {
					panic(err)
				}

				//m.Log(rpc.Output().Package().ProtoName())

				rpcModel := rpcModel{RpcName: rpc.Name().UpperCamelCase().String(), Input: rpc.Input().Name().UpperCamelCase().String(), Output: rpc.Output().Name().UpperCamelCase().String(), Option: rights,PackageName: m.Context.PackageName(f).String()}

				re := regexp.MustCompile("{([^{]*)}")
				fieldsInResource := re.FindAllStringSubmatch(rights.Resource, -1)
				//m.Log(rights)
				fieldVsfieldSeperated := make(map[string][]string, 0)
				fieldVsGetString := make(map[string]string, 0)
				for _, fr := range fieldsInResource {
					//m.Log(fr[1])
					fieldVsfieldSeperated[fr[1]] = strings.Split(fr[1], ".")
				}

				//m.Log(inputField.Name())
				for fr, dotSeperatedkeys := range fieldVsfieldSeperated {

					for _, r := range dotSeperatedkeys {
						if _, ok := fieldVsGetString[fr]; ok {
							fieldVsGetString[fr] = fieldVsGetString[fr] + ".Get" + toCamelInitCase(r,true)+"()"
						} else {
							fieldVsGetString[fr] = "Get" + toCamelInitCase(r,true)+"()"
						}
					}
					if fieldVsGetString[fr] != "" {
						//m.Log(fieldVsGetString[fr],fr)
						rpcModel.GetString = append(rpcModel.GetString, fieldVsGetString[fr])
					}
				}

				service.Rpcs = append(service.Rpcs, rpcModel)
			}
			fm.Services = append(fm.Services, service)
			//for _, msg := range f.AllMessages() {
			//
			//	fields := msg.Fields()
			//	mp := messageModel{}
			//	mp.Name = msg.Name().UpperCamelCase().String()
			//
			//	list := make([]zapField, len(f.AllMessages()))
			//
			//	for _, v := range fields {
			//
			//		redact := false
			//		_, err := v.Extension(zappb.E_Redact, &redact)
			//		if err != nil {
			//			m.Log(err)
			//		}
			//
			//		r := zapField{
			//			Redact:   redact,
			//			Name:     v.Name().UpperCamelCase().String(),
			//			Type:     v.Descriptor().Type.String(),
			//			Label:    v.Descriptor().GetLabel().String(),
			//			TypeName: v.Descriptor().GetTypeName(),
			//		}
			//		list = append(list, r)
			//
			//		//m.Log(r)
			//
			//	}
			//	mp.Fields = list
			//	fm.Messages = append(fm.Messages, mp)
			//}
		}



		m.OverwriteGeneratorTemplateFile(
			name,
			T.Lookup("File"),
			&fm,
		)
	}

	return m.Artifacts()

}

type rpcModel struct {
	PackageName string
	RpcName     string
	Input    string
	Output   string
	GetString []string
	Option   zappb.MyRights
}

type serviceModel struct {
	ServiceName   string
	PackageName string
	Rpcs []rpcModel
}

type fileModel struct {
	PackageName string
	Imports     []string
	Services    []serviceModel
}
// Converts a string to CamelCase
func toCamelInitCase(s string, initCase bool) string {
	s = addWordBoundariesToNumbers(s)
	s = strings.Trim(s, " ")
	n := ""
	capNext := initCase
	for _, v := range s {
		if v >= 'A' && v <= 'Z' {
			n += string(v)
		}
		if v >= '0' && v <= '9' {
			n += string(v)
		}
		if v >= 'a' && v <= 'z' {
			if capNext {
				n += strings.ToUpper(string(v))
			} else {
				n += string(v)
			}
		}
		if v == '_' || v == ' ' || v == '-' {
			capNext = true
		} else {
			capNext = false
		}
	}
	return n
}
var numberSequence = regexp.MustCompile(`([a-zA-Z])(\d+)([a-zA-Z]?)`)
var numberReplacement = []byte(`$1 $2 $3`)

func addWordBoundariesToNumbers(s string) string {
	b := []byte(s)
	b = numberSequence.ReplaceAll(b, numberReplacement)
	return string(b)
}