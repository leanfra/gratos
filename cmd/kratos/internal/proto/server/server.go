package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CmdServer the service command.
var CmdServer = &cobra.Command{
	Use:   "server",
	Short: "Generate the proto server implementations",
	Long:  "Generate the proto server implementations. Example: kratos proto server api/xxx.proto --target-dir=internal/service",
	Run:   run,
}
var targetDir string
var bizDir string
var dataDir string
var modelFile string
var modelComment string
var overWrite bool
var nobiz bool
var nobizmodel bool
var noservice bool
var nodata bool

func init() {
	CmdServer.Flags().StringVarP(&targetDir, "service-dir", "t", "internal/service", "generated service directory. one file per service")
	CmdServer.Flags().StringVarP(&bizDir, "biz-dir", "b", "internal/biz", "generated biz directory. one file per service")
	CmdServer.Flags().StringVarP(&dataDir, "data-dir", "d", "internal/data", "generated data directory. one file per service")
	CmdServer.Flags().StringVarP(&modelFile, "model-file", "m", "models.go", "generated model file under biz directory")
	CmdServer.Flags().StringVarP(&modelComment, "model-comment", "c", "gratos::model", "comment tag to message converted to model")
	CmdServer.Flags().BoolVarP(&overWrite, "over-write", "f", false, "force over write existed file")
	CmdServer.Flags().BoolVarP(&nobiz, "no-biz", "", false, "not generate biz code")
	CmdServer.Flags().BoolVarP(&nobizmodel, "no-biz-model", "", false, "not generate biz model code")
	CmdServer.Flags().BoolVarP(&noservice, "no-service", "", false, "not generate service code")
	CmdServer.Flags().BoolVarP(&nodata, "no-data", "", false, "not generate data code")
}

func run(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify the proto file. Example: kratos proto server api/xxx.proto")
		return
	}
	reader, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	var (
		pkg      string
		services []*Service
	)
	models := Models{}
	proto.Walk(definition,
		proto.WithOption(func(o *proto.Option) {
			if o.Name == "go_package" {
				pkg = strings.Split(o.Constant.Source, ";")[0]
			}
		}),
		proto.WithService(func(s *proto.Service) {
			project := strings.Split(pkg, "/")[0]
			cs := &Service{
				Project: project,
				Package: pkg,
				Service: serviceName(s.Name),
			}
			for _, e := range s.Elements {
				r, ok := e.(*proto.RPC)
				if !ok {
					continue
				}
				cs.Methods = append(cs.Methods, &Method{
					Service: serviceName(s.Name), Name: serviceName(r.Name), Request: parametersName(r.RequestType),
					Reply: parametersName(r.ReturnsType), Type: getMethodType(r.StreamsRequest, r.StreamsReturns),
				})
			}
			services = append(services, cs)
		}),
		proto.WithMessage(func(m *proto.Message) {
			//fmt.Println(m.Name, m.Comment)
			is_model := false
			if m.Comment != nil {
				comments := m.Comment.Lines
				for _, c := range comments {
					if strings.Contains(c, modelComment) {
						is_model = true
					}
				}
				if !is_model {
					return
				}
				// valid module
				_model := Model{
					Name: m.Name,
				}
				for _, e := range m.Elements {
					nf, ok := e.(*proto.NormalField)
					if ok {
						//fmt.Println(nf.Name, nf.Type)
						_model.Fields = append(_model.Fields, Field{
							Name:     fieldName(nf.Name),
							Type:     nf.Type,
							Repeated: nf.Repeated,
						})
					}
				}
				models.Models = append(models.Models, _model)
			}
		}),
	)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Target directory: %s does not exist\n", targetDir)
		return
	}
	//fmt.Println(models)
	if !nobizmodel {
		model_to := filepath.Join(targetDir, "../biz", modelFile)
		biz_model_bytes, err := models.execute()
		if err != nil {
			log.Fatal(err)
		}
		if e := writeFile(model_to, biz_model_bytes); e == nil {
			fmt.Printf("generate: %s\n", model_to)
		}
	}

	for _, s := range services {

		// service
		if !noservice {
			svc_file := filepath.Join(targetDir, strings.ToLower(s.Service)+".go")
			svc_bytes, err := s.execute()
			if err != nil {
				log.Fatal(err)
			}
			if e := writeFile(svc_file, svc_bytes); e == nil {
				fmt.Printf("generate: %s\n", svc_file)
			}
		}

		// biz
		if !nobiz {
			biz_bytes, err := s.executeBiz()
			if err != nil {
				log.Fatal(err)
			}
			biz_file := filepath.Join(bizDir, strings.ToLower(s.Service)+".go")
			if e := writeFile(biz_file, biz_bytes); e == nil {
				fmt.Printf("generate: %s\n", biz_file)
			}
		}

		// data
		if !nodata {
			data_bytes, err := s.executeData()
			if err != nil {
				log.Fatal(err)
			}
			data_file := filepath.Join(dataDir, strings.ToLower(s.Service)+".go")
			if e := writeFile(data_file, data_bytes); e == nil {
				fmt.Printf("generate: %s\n", data_file)
			}
		}
	}
}

func writeFile(to string, data []byte) error {
	if _, err := os.Stat(to); !os.IsNotExist(err) {
		if overWrite {
			fmt.Printf("over write: %s\n", to)
		} else {
			fmt.Fprintf(os.Stderr, "already exists: %s\n", to)
			return fmt.Errorf("file already exists: %s", to)
		}
	}
	if err := os.WriteFile(to, data, 0o644); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func getMethodType(streamsRequest, streamsReturns bool) MethodType {
	if !streamsRequest && !streamsReturns {
		return unaryType
	} else if streamsRequest && streamsReturns {
		return twoWayStreamsType
	} else if streamsRequest {
		return requestStreamsType
	} else if streamsReturns {
		return returnsStreamsType
	}
	return unaryType
}

func parametersName(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

func serviceName(name string) string {
	return toUpperCamelCase(strings.Split(name, ".")[0])
}

func fieldName(name string) string {
	return toUpperCamelCase(strings.Split(name, ".")[0])
}

func toUpperCamelCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = cases.Title(language.Und, cases.NoLower).String(s)
	return strings.ReplaceAll(s, " ", "")
}
