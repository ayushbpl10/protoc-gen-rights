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
				option, err := proto.GetExtension(opt, rightspb.E_Validator)
				if err != nil {
					panic(err)
				}
				byteData, err := json.Marshal(option)
				if err != nil {
					panic(err)
				}
				right := rightspb.MyRights{}
				err = json.Unmarshal(byteData, &right)
				if err != nil {
					panic(err)
				}

				//m.Log(rpc.Output().Package().ProtoName())

				rpcModel := rpcModel{RpcName: rpc.Name().UpperCamelCase().String(), Input: rpc.Input().Name().UpperCamelCase().String(), Output: rpc.Output().Name().UpperCamelCase().String(), Option: right,PackageName: m.Context.PackageName(f).String()}

				re := regexp.MustCompile("{([^{]*)}")
				//[right] : [with{} , without{}]
				fieldsInResource := make(map[string][][]string,0)
				for _, fieldRight := range right.Resource {
					fieldsInResource[fieldRight] = append(fieldsInResource[fieldRight],re.FindAllStringSubmatch(fieldRight, -1)...)
				}

				//[right] : {[without {}] : seperatedFields}
				fieldVsfieldSeperatedRightMap := make(map[string]map[string][]string,0)

				//Track keys to maintan map order
				fieldVsfieldSeperatedRightMapTrack := make(map[string][]string)

				ToBEreplacedByPlaceHolder := make([]string,0)

				for fieldRight, ArrayOfCurlyBraces := range fieldsInResource {
					fieldVsfieldSeperated := make(map[string][]string, 0)
					for _, fr := range ArrayOfCurlyBraces {
						//%s place holders
						ToBEreplacedByPlaceHolder = append(ToBEreplacedByPlaceHolder, fr[0])
						//splitting the dot operator
						fieldVsfieldSeperated[fr[1]] = strings.Split(fr[1], ".")

						fieldVsfieldSeperatedRightMapTrack[fieldRight] = append(fieldVsfieldSeperatedRightMapTrack[fieldRight],fr[1])
						fieldVsfieldSeperatedRightMap[fieldRight] = fieldVsfieldSeperated
					}
				}

				//[right] : isrepeated
				RightRepeatedMap := make(map[string]bool)

				for rightVal, fieldVsfieldSeperatedElement := range fieldVsfieldSeperatedRightMap {

					for _,inOrder := range fieldVsfieldSeperatedRightMapTrack[rightVal]{

						dotSeperatedkeys,_  := fieldVsfieldSeperatedElement[inOrder]

						for _, r := range dotSeperatedkeys {
							//m.Log(r)
							//checking the key is a field in message
							found := false
							IsRepeated := false
							for _,msg := range f.AllMessages() {
								for _,field := range msg.Fields() {
									if field.Name().String() == r {
										found = true
										IsRepeated = field.Type().IsRepeated()
									}
								}
							}
							if !found {
								m.Log("key is not a primitive type : " + r)
							}

							if IsRepeated {
								RightRepeatedMap[rightVal] = true
							}
						}
					}


				}


				//if GlobalIsRepeated {
					for rightVal, fieldVsfieldSeperatedElement := range fieldVsfieldSeperatedRightMap {

						fieldVsGetString := make(map[string]string, 0)
						resource := Resource{}
						//if repeated
						if RightRepeatedMap[rightVal] {
							for _,inOrder := range fieldVsfieldSeperatedRightMapTrack[rightVal]{
								fr := inOrder
								dotSeperatedkeys,_ := fieldVsfieldSeperatedElement[inOrder]

									//m.Log(fr)
									//[without curly braces] : Current For Loop Dotted Access
									ForLoopMapWithGet := make(map[string]string, 0)
									ForLoopMapWithOutGet := make(map[string]string, 0)
									for _, r := range dotSeperatedkeys {

										//checking field is repeated
										IsRepeated := false
										IsNotPrimitive := false
										IsRepeatedPrimitive := false
										for _, msg := range f.AllMessages() {

											for _, field := range msg.Fields() {
												if field.Name().String() == r {
													IsRepeated = field.Type().IsRepeated()
													IsNotPrimitive = field.Type().IsEmbed()
													if field.Type().Element() != nil {
														IsRepeatedPrimitive = field.Type().Element().IsEmbed()
													}

												}
											}
										}

										if IsRepeated {

											if _, ok := ForLoopMapWithGet[fr]; !ok {
												ForLoopMapWithGet[fr] = "Get" + toCamelInitCase(r, true)
												ForLoopMapWithOutGet[fr] = toCamelInitCase(r, true)

												//for loops already contains the previous repeated value
												for _,forLoopResourceExists := range resource.ForLoop {
													if strings.Contains(forLoopResourceExists.RangeKey, ForLoopMapWithOutGet[fr]) {
														resource.ForLoop = resource.ForLoop[0 : len(resource.ForLoop)-1]
													}
												}

											} else {

												//for loops already contains the previous repeated value
												if strings.Contains(fr, ForLoopMapWithOutGet[fr]) {
													resource.ForLoop = resource.ForLoop[0 : len(resource.ForLoop)-1]
												}
												ForLoopMapWithGet[fr] = ForLoopMapWithOutGet[fr] + "." + "Get" + toCamelInitCase(r, true)
											}


											forLoop := ForLoop{}
											forLoop.RangeKey = ForLoopMapWithGet[fr]
											forLoop.ValueKey = toCamelInitCase(r, true)

											resource.ForLoop = append(resource.ForLoop, forLoop)

											//starting the getString from inner for loop value
											fieldVsGetString[fr] = toCamelInitCase(r, true)

										} else {
											if _, ok := fieldVsGetString[fr]; ok {
												fieldVsGetString[fr] = fieldVsGetString[fr] + ".Get" + toCamelInitCase(r, true) + "()"
											} else {
												fieldVsGetString[fr] = "Get" + toCamelInitCase(r, true) + "()"
											}
										}

										resource.IsRepeated = RightRepeatedMap[rightVal]

										if !IsRepeated {
											if !IsNotPrimitive {
												found := false
												for _,strMap := range resource.GetStrings {
													if _,ok := strMap[fieldVsGetString[fr]]; ok {
															found = true
													}
												}
												if !found {
													mapGetString := make(map[string]bool, 0)
													mapGetString[fieldVsGetString[fr]] = false
													for _,forLoop := range resource.ForLoop {
														if strings.Contains(fieldVsGetString[fr],forLoop.ValueKey){
															//m.Log(fr,fieldVsGetString[fr])
															mapGetString[fieldVsGetString[fr]] = true
														}
													}

													resource.GetStrings = append(resource.GetStrings, mapGetString)
												}
											}
										}else{
											if !IsRepeatedPrimitive {
												mapGetString := make(map[string]bool, 0)
												mapGetString[fieldVsGetString[fr]] = false
												for _,forLoop := range resource.ForLoop {
													if strings.Contains(fieldVsGetString[fr],forLoop.ValueKey){
														//m.Log(fr,fieldVsGetString[fr])
														mapGetString[fieldVsGetString[fr]] = true
													}
												}

												resource.GetStrings = append(resource.GetStrings, mapGetString)
											}

										}

										resource.ResourceStringWithCurlyBraces = rightVal
									}



							}
							//preparing formatted string
							resource.ResourceStringWithFormatter = rightVal
							for _ , p := range ToBEreplacedByPlaceHolder {
								resource.ResourceStringWithFormatter = strings.Replace(resource.ResourceStringWithFormatter,p,"%s",-1)
							}

							rpcModel.Resources = append(rpcModel.Resources, resource)

						} else{
							for fr, dotSeperatedkeys := range fieldVsfieldSeperatedElement {

								for _, r := range dotSeperatedkeys {

									//checking field is repeated
									IsNotPrimitive := false
									for _,msg := range f.AllMessages() {

										for _,field := range msg.Fields() {
											if field.Name().String() == r {
												IsNotPrimitive = field.Type().IsEmbed()
											}
										}
									}

									if _, ok := fieldVsGetString[fr]; ok {
										fieldVsGetString[fr] = fieldVsGetString[fr] + ".Get" + toCamelInitCase(r,true)+"()"
									} else {
										fieldVsGetString[fr] = "Get" + toCamelInitCase(r,true)+"()"
									}


									resource.IsRepeated = RightRepeatedMap[rightVal]

									if !IsNotPrimitive {
										mapGetString := make(map[string]bool, 0)
										mapGetString[fieldVsGetString[fr]] = false
										resource.GetStrings = append(resource.GetStrings,mapGetString)
									}

									resource.ResourceStringWithCurlyBraces = rightVal
								}

							}

							//preparing formatted string
							resource.ResourceStringWithFormatter = rightVal
							for _ , p := range ToBEreplacedByPlaceHolder {
								resource.ResourceStringWithFormatter = strings.Replace(resource.ResourceStringWithFormatter,p,"%s",-1)
							}

							rpcModel.Resources = append(rpcModel.Resources, resource)
						}
					}

					service.Rpcs = append(service.Rpcs, rpcModel)
				}

				fm.Services = append(fm.Services, service)
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
	Input       string
	Output      string
	Option      rightspb.MyRights
	Resources   []Resource
}

type Resource struct {
	IsRepeated  					bool
	GetStrings   					[]map[string]bool
	ResourceStringWithCurlyBraces 	string
	ResourceStringWithFormatter     string
	ForLoop     					[]ForLoop
}

type ForLoop struct {
	RangeKey 	   string
	ValueKey string
	Level      int
}

type serviceModel struct {
	ServiceName   string
	PackageName   string
	Rpcs          []rpcModel
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

func reverseString(input []string) []string {
	if len(input) == 0 {
		return input
	}
	return append(reverseString(input[1:]), input[0])
}

func ReturnFields(field pgs.Field, fields []pgs.Field) []pgs.Field {

	if field.Type().IsEmbed() {
		fields = append(fields,ReturnFields(field, fields)...)
	}
	fields = append(fields, field)
	return fields
}