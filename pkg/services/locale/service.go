package locale

import (
	"fmt"
	"log/slog"

	"github.com/gravestench/servicemesh"
	tbl "github.com/gravestench/tbl_text"

	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/tblLoader"
	"github.com/gravestench/dark-magic/pkg/services/webRouter"
)

type recipe interface {
	servicemesh.Service
	servicemesh.HasLogger
	servicemesh.HasDependencies
	webRouter.HasRouteSlug
	webRouter.IsRouteInitializer
	LoadsStringTables
}

type Service struct {
	logger       *slog.Logger
	mpq          mpqLoader.Dependency
	tbl          tblLoader.Dependency
	language     string
	stringTables struct {
		Vanilla   tbl.TextTable
		Expansion tbl.TextTable
		Patch     tbl.TextTable
	}
	compositeStringTable tbl.TextTable
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.compositeStringTable = make(tbl.TextTable)

	languages := s.GetSupportedLanguages()

	if len(languages) < 1 {
		s.logger.Error("no locale languages found")
	}

	s.language = languages[0]
	s.loadTextTables()
}

func (s *Service) Name() string {
	return "Locale"
}

func (s *Service) Ready() bool {
	if s.mpq == nil {
		return false
	}

	if s.tbl == nil {
		return false
	}

	return true
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}

func (s *Service) GetSupportedLanguages() []string {
	code := s.getLanguageCodeFromMpq()
	l := GetLanguageLiteral(code)

	return []string{l}
}

func (s *Service) GetSupportedCharsets() []string {
	//TODO implement me
	panic("implement me")
}

func (s *Service) GetTextByKey(key string) (string, error) {
	val, found := s.compositeStringTable[key]
	if !found {
		return key, fmt.Errorf("no value found for key %q", key)
	}

	return val, nil
}

func (s *Service) getLanguageCodeFromMpq() byte {
	return 0
}

func (s *Service) loadTextTables() {
	localePathExpansionStringTable := fmt.Sprintf(ExpansionStringTable, s.language)
	localePathStringTable := fmt.Sprintf(StringTable, s.language)
	localePathPatchStringTable := fmt.Sprintf(PatchStringTable, s.language)

	if tbl, err := s.tbl.Load(localePathStringTable); err != nil {
		s.logger.Error("loading Vanilla string table")
	} else {
		s.stringTables.Vanilla = tbl
	}

	if tbl, err := s.tbl.Load(localePathExpansionStringTable); err != nil {
		s.logger.Error("loading Expansion string table")
	} else {
		s.stringTables.Expansion = tbl
	}

	if tbl, err := s.tbl.Load(localePathPatchStringTable); err != nil {
		s.logger.Error("loading Patch string table")
	} else {
		s.stringTables.Patch = tbl
	}

	for key, value := range s.stringTables.Vanilla {
		s.compositeStringTable[key] = value
	}

	for key, value := range s.stringTables.Expansion {
		s.compositeStringTable[key] = value
	}

	for key, value := range s.stringTables.Patch {
		s.compositeStringTable[key] = value
	}
}
