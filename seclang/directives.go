// Copyright 2021 Juan Pablo Tosso
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package seclang

import (
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"strconv"
	"strings"

	"github.com/jptosso/coraza-waf/v2"
	"github.com/jptosso/coraza-waf/v2/types"
	"go.uber.org/zap"
)

type directive = func(w *coraza.Waf, opts string) error

func RegisterDirectivePlugin(name string, directive func(w *coraza.Waf, opts string) error) {
	directivesMap[strings.ToLower(name)] = directive
}

func directiveSecComponentSignature(w *coraza.Waf, opts string) error {
	w.ComponentNames = append(w.ComponentNames, opts)
	return nil
}

func directiveSecMarker(w *coraza.Waf, opts string) error {
	rule, err := ParseRule(RuleOptions{
		Waf:          w,
		Data:         "id:1, pass, nolog",
		WithOperator: false,
	})
	if err != nil {
		return err
	}
	rule.SecMark = opts
	rule.ID = 0
	rule.Phase = 0
	if err := w.Rules.Add(rule); err != nil {
		if perr := fmt.Errorf("Failed to compile rule (%s): %s", err, opts); perr != nil {
			return perr // can't write to log, return this instead
		}
		return err
	}
	w.Logger.Debug("added secmark rule")
	return nil
}

func directiveSecAction(w *coraza.Waf, opts string) error {
	rule, err := ParseRule(RuleOptions{
		Waf:          w,
		Data:         opts,
		WithOperator: false,
	})
	if err != nil {
		if perr := fmt.Errorf("Failed to compile rule (%s): %s", err, opts); perr != nil {
			return perr // can't write to log, return this instead
		}
		return err
	}
	if err := w.Rules.Add(rule); err != nil {
		if perr := fmt.Errorf("Failed to compile rule (%s): %s", err, opts); perr != nil {
			return perr // can't write to log, return this instead
		}
		return err
	}
	w.Logger.Debug("Added SecAction",
		zap.String("rule", opts),
	)
	return nil
}

func directiveSecRule(w *coraza.Waf, opts string) error {
	line, _ := w.GetConfig("parser_last_line", 0).(int)
	configFile, _ := w.GetConfig("parser_config_file", "").(string)
	configDir, _ := w.GetConfig("parser_config_dir", "").(string)
	rule, err := ParseRule(RuleOptions{
		Waf:          w,
		Data:         opts,
		WithOperator: true,
		Line:         line,
		ConfigFile:   configFile,
		ConfigDir:    configDir,
	})
	if err != nil {
		if perr := fmt.Errorf("Failed to compile rule (%s): %s", err, opts); perr != nil {
			return perr // can't write to log, return this instead
		}
		return err
	}
	return w.Rules.Add(rule)
}

func directiveSecResponseBodyAccess(w *coraza.Waf, opts string) error {
	w.ResponseBodyAccess = (strings.ToLower(opts) == "on")
	return nil
}

func directiveSecRequestBodyLimit(w *coraza.Waf, opts string) error {
	limit, _ := strconv.ParseInt(opts, 10, 64)
	w.RequestBodyLimit = limit
	return nil
}

func directiveSecRequestBodyAccess(w *coraza.Waf, opts string) error {
	w.RequestBodyAccess = (strings.ToLower(opts) == "on")
	return nil
}

func directiveSecRuleEngine(w *coraza.Waf, opts string) error {
	engine, err := types.ParseRuleEngineStatus(opts)
	w.RuleEngine = engine
	return err
}

func directiveUnsupported(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecWebAppID(w *coraza.Waf, opts string) error {
	w.WebAppID = opts
	return nil
}

func directiveSecTmpDir(w *coraza.Waf, opts string) error {
	w.TmpDir = opts
	return nil
}

func directiveSecServerSignature(w *coraza.Waf, opts string) error {
	w.ServerSignature = opts
	return nil
}

func directiveSecRuleRemoveByTag(w *coraza.Waf, opts string) error {
	for _, r := range w.Rules.FindByTag(opts) {
		w.Rules.DeleteByID(r.ID)
	}
	return nil
}

func directiveSecRuleRemoveByMsg(w *coraza.Waf, opts string) error {
	for _, r := range w.Rules.FindByMsg(opts) {
		w.Rules.DeleteByID(r.ID)
	}
	return nil
}

func directiveSecRuleRemoveByID(w *coraza.Waf, opts string) error {
	id, _ := strconv.Atoi(opts)
	w.Rules.DeleteByID(id)
	return nil
}

func directiveSecResponseBodyMimeTypesClear(w *coraza.Waf, opts string) error {
	w.ResponseBodyMimeTypes = []string{}
	return nil
}

func directiveSecResponseBodyMimeType(w *coraza.Waf, opts string) error {
	w.ResponseBodyMimeTypes = strings.Split(opts, " ")
	return nil
}

func directiveSecResponseBodyLimitAction(w *coraza.Waf, opts string) error {
	w.RejectOnResponseBodyLimit = (strings.ToLower(opts) == "reject")
	return nil
}

func directiveSecResponseBodyLimit(w *coraza.Waf, opts string) error {
	var err error
	w.ResponseBodyLimit, err = strconv.ParseInt(opts, 10, 64)
	return err
}

func directiveSecRequestBodyLimitAction(w *coraza.Waf, opts string) error {
	w.RejectOnRequestBodyLimit = (strings.ToLower(opts) == "reject")
	return nil
}

func directiveSecRequestBodyInMemoryLimit(w *coraza.Waf, opts string) error {
	w.RequestBodyInMemoryLimit, _ = strconv.ParseInt(opts, 10, 64)
	return nil
}

func directiveSecRemoteRulesFailAction(w *coraza.Waf, opts string) error {
	w.AbortOnRemoteRulesFail = (strings.ToLower(opts) == "abort")
	return nil
}

func directiveSecRemoteRules(w *coraza.Waf, opts string) error {
	return fmt.Errorf("not implemented")
}

func directiveSecConnWriteStateLimit(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecSensorID(w *coraza.Waf, opts string) error {
	w.SensorID = opts
	return nil
}

func directiveSecConnReadStateLimit(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecPcreMatchLimitRecursion(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecPcreMatchLimit(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecHTTPBlKey(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecGsbLookupDb(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecHashMethodPm(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecHashMethodRx(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecHashParam(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecHashKey(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecHashEngine(w *coraza.Waf, opts string) error {
	return nil
}

func directiveSecDefaultAction(w *coraza.Waf, opts string) error {
	da, ok := w.GetConfig("rule_default_actions", []string{}).([]string)
	if !ok {
		da = []string{}
	}
	da = append(da, opts)
	w.SetConfig("rule_default_actions", da)
	return nil
}

func directiveSecContentInjection(w *coraza.Waf, opts string) error {
	w.ContentInjection = parseBoolean(opts)
	return nil
}

func directiveSecConnEngine(w *coraza.Waf, opts string) error {
	/*
		switch opts{
		case "On":
			w.ConnEngine = engine.CONN_ENGINE_ON
			break
		case "Off":
			w.ConnEngine = engine.CONN_ENGINE_OFF
			break
		case "DetectOnly":
			w.ConnEngine = engine.CONN_ENGINE_DETECTONLY
			break
		}
		break
	*/
	return nil
}

func directiveSecCollectionTimeout(w *coraza.Waf, opts string) error {
	// w.CollectionTimeout, _ = strconv.Atoi(opts)
	return nil
}

func directiveSecAuditLog(w *coraza.Waf, opts string) error {
	if len(opts) == 0 {
		return errors.New("syntax error: SecAuditLog /some/absolute/path.log")
	}
	w.AuditLog = opts
	if err := w.UpdateAuditLogger(); err != nil {
		return err
	}
	return nil
}

func directiveSecAuditLogType(w *coraza.Waf, opts string) error {
	if len(opts) == 0 {
		return errors.New("syntax error: SecAuditLogType [concurrent/https/serial/...]")
	}
	w.AuditLogType = strings.ToLower(opts)
	if err := w.UpdateAuditLogger(); err != nil {
		return err
	}
	return nil
}

func directiveSecAuditLogFormat(w *coraza.Waf, opts string) error {
	if len(opts) == 0 {
		return errors.New("syntax error: SecAuditLogFormat [json/jsonlegacy/native/...]")
	}
	w.AuditLogFormat = strings.ToLower(opts)
	if err := w.UpdateAuditLogger(); err != nil {
		return err
	}
	return nil
}

func directiveSecAuditLogDir(w *coraza.Waf, opts string) error {
	if len(opts) == 0 {
		return errors.New("syntax error: SecAuditLogDir /some/absolute/path")
	}
	w.AuditLogDir = opts
	if err := w.UpdateAuditLogger(); err != nil {
		return err
	}
	return nil
}

func directiveSecAuditLogDirMode(w *coraza.Waf, opts string) error {
	if len(opts) == 0 {
		return errors.New("syntax error: SecAuditLogDirMode [0777/0700/...]")
	}
	// w.AuditLogDirMode, _ = strconv.ParseInt(opts, 8, 32)
	if err := w.UpdateAuditLogger(); err != nil {
		return err
	}
	return nil
}

func directiveSecAuditLogFileMode(w *coraza.Waf, opts string) error {
	if len(opts) == 0 {
		return errors.New("syntax error: SecAuditLogFileMode [0777/0700/...]")
	}
	// w.AuditLogFileMode, _ = strconv.ParseInt(opts, 8, 32)
	if err := w.UpdateAuditLogger(); err != nil {
		return err
	}
	return nil
}

func directiveSecAuditLogRelevantStatus(w *coraza.Waf, opts string) error {
	var err error
	w.AuditLogRelevantStatus, err = regexp.Compile(opts)
	return err
}

func directiveSecAuditLogParts(w *coraza.Waf, opts string) error {
	w.AuditLogParts = types.AuditLogParts(opts)
	return nil
}

func directiveSecAuditEngine(w *coraza.Waf, opts string) error {
	au, err := types.ParseAuditEngineStatus(opts)
	w.AuditEngine = au
	return err
}

func directiveSecDataDir(w *coraza.Waf, opts string) error {
	// TODO validations
	w.DataDir = opts
	return nil
}

func directiveSecUploadKeepFiles(w *coraza.Waf, opts string) error {
	w.UploadKeepFiles = parseBoolean(opts)
	return nil
}

func directiveSecUploadFileMode(w *coraza.Waf, opts string) error {
	fm, err := strconv.ParseInt(opts, 8, 32)
	w.UploadFileMode = fs.FileMode(fm)
	return err
}

func directiveSecUploadFileLimit(w *coraza.Waf, opts string) error {
	var err error
	w.UploadFileLimit, err = strconv.Atoi(opts)
	return err
}

func directiveSecUploadDir(w *coraza.Waf, opts string) error {
	// TODO validations
	w.UploadDir = opts
	return nil
}

func directiveSecRequestBodyNoFilesLimit(w *coraza.Waf, opts string) error {
	var err error
	w.RequestBodyNoFilesLimit, err = strconv.ParseInt(opts, 10, 64)
	return err
}

func directiveSecDebugLog(w *coraza.Waf, opts string) error {
	return w.SetDebugLogPath(opts)
}

func directiveSecDebugLogLevel(w *coraza.Waf, opts string) error {
	lvl, err := strconv.Atoi(opts)
	if err != nil {
		return err
	}
	return w.SetLogLevel(lvl)
}

func directiveSecRuleUpdateTargetById(w *coraza.Waf, opts string) error {
	spl := strings.SplitN(opts, " ", 2)
	if len(spl) != 2 {
		return errors.New("syntax error: SecRuleUpdateTargetById id \"VARIABLES\"")
	}
	id, err := strconv.Atoi(spl[0])
	if err != nil {
		return err
	}
	rule := w.Rules.FindByID(id)
	rp := &RuleParser{
		rule:           rule,
		options:        RuleOptions{},
		defaultActions: map[types.RulePhase][]ruleAction{},
	}
	return rp.ParseVariables(strings.Trim(spl[1], "\""))
}

func parseBoolean(data string) bool {
	data = strings.ToLower(data)
	return data == "on"
}

var (
	_ directive = directiveSecAction
	_ directive = directiveSecAuditEngine
	_ directive = directiveSecAuditLog
	_ directive = directiveSecAuditLogType
	_ directive = directiveSecAuditLogFormat
	_ directive = directiveSecAuditLogParts
	_ directive = directiveSecAuditLogRelevantStatus
	_ directive = directiveSecContentInjection
	_ directive = directiveSecDataDir
	_ directive = directiveSecDefaultAction
	_ directive = directiveSecDebugLog
	_ directive = directiveSecDebugLogLevel
	_ directive = directiveSecHashEngine
	_ directive = directiveSecHashKey
	_ directive = directiveSecHashMethodPm
	_ directive = directiveSecHashMethodRx
	_ directive = directiveSecHashParam
	_ directive = directiveSecHTTPBlKey
	_ directive = directiveSecMarker
	_ directive = directiveSecRemoteRules
	_ directive = directiveSecSensorID
	_ directive = directiveSecRuleUpdateTargetById
)

var directivesMap = map[string]directive{
	"secwebappid":                   directiveSecWebAppID,
	"secuploadkeepfiles":            directiveSecUploadKeepFiles,
	"secuploadfilemode":             directiveSecUploadFileMode,
	"secuploadfilelimit":            directiveSecUploadFileLimit,
	"secuploaddir":                  directiveSecUploadDir,
	"sectmpdir":                     directiveSecTmpDir,
	"secserversignature":            directiveSecServerSignature,
	"secsensorid":                   directiveSecSensorID,
	"secruleremovebytag":            directiveSecRuleRemoveByTag,
	"secruleremovebymsg":            directiveSecRuleRemoveByMsg,
	"secruleremovebyid":             directiveSecRuleRemoveByID,
	"secruleengine":                 directiveSecRuleEngine,
	"secrule":                       directiveSecRule,
	"secresponsebodymimetypesclear": directiveSecResponseBodyMimeTypesClear,
	"secresponsebodymimetype":       directiveSecResponseBodyMimeType,
	"secresponsebodylimitaction":    directiveSecResponseBodyLimitAction,
	"secresponsebodylimit":          directiveSecResponseBodyLimit,
	"secresponsebodyaccess":         directiveSecResponseBodyAccess,
	"secrequestbodynofileslimit":    directiveSecRequestBodyNoFilesLimit,
	"secrequestbodylimitaction":     directiveSecRequestBodyLimitAction,
	"secrequestbodylimit":           directiveSecRequestBodyLimit,
	"secrequestbodyinmemorylimit":   directiveSecRequestBodyInMemoryLimit,
	"secrequestbodyaccess":          directiveSecRequestBodyAccess,
	"secremoterulesfailaction":      directiveSecRemoteRulesFailAction,
	"secremoterules":                directiveSecRemoteRules,
	"secpcrematchlimitrecursion":    directiveSecPcreMatchLimitRecursion,
	"secpcrematchlimit":             directiveSecPcreMatchLimit,
	"secmarker":                     directiveSecMarker,
	"sechttpblkey":                  directiveSecHTTPBlKey,
	"sechashparam":                  directiveSecHashParam,
	"sechashmethodrx":               directiveSecHashMethodRx,
	"sechashmethodpm":               directiveSecHashMethodPm,
	"sechashkey":                    directiveSecHashKey,
	"sechashengine":                 directiveSecHashEngine,
	"secgsblookupdb":                directiveSecGsbLookupDb,
	"secdefaultaction":              directiveSecDefaultAction,
	"secdatadir":                    directiveSecDataDir,
	"seccontentinjection":           directiveSecContentInjection,
	"secconnwritestatelimit":        directiveSecConnWriteStateLimit,
	"secconnreadstatelimit":         directiveSecConnReadStateLimit,
	"secconnengine":                 directiveSecConnEngine,
	"seccomponentsignature":         directiveSecComponentSignature,
	"seccollectiontimeout":          directiveSecCollectionTimeout,
	"secauditlogrelevantstatus":     directiveSecAuditLogRelevantStatus,
	"secauditlogparts":              directiveSecAuditLogParts,
	"secauditlogdir":                directiveSecAuditLogDir,
	"secauditlogstoragedir":         directiveSecAuditLogDir,
	"secauditlog":                   directiveSecAuditLog,
	"secauditengine":                directiveSecAuditEngine,
	"secaction":                     directiveSecAction,
	"secdebuglog":                   directiveSecDebugLog,
	"secdebugloglevel":              directiveSecDebugLogLevel,
	"secauditlogformat":             directiveSecAuditLogFormat,
	"secauditlogtype":               directiveSecAuditLogType,
	"secauditlogfilemode":           directiveSecAuditLogFileMode,
	"secauditlogdirmode":            directiveSecAuditLogDirMode,

	// Unsupported Directives
	"secargumentseparator":     directiveUnsupported,
	"seccookieformat":          directiveUnsupported,
	"secruleupdatetargetbytag": directiveUnsupported,
	"secruleupdatetargetbymsg": directiveUnsupported,
	"secruleupdatetargetbyid":  directiveSecRuleUpdateTargetById,
	"secruleupdateactionbyid":  directiveUnsupported,
	"secrulescript":            directiveUnsupported,
	"secruleperftime":          directiveUnsupported,
	"SecUnicodeMap":            directiveUnsupported,
}
