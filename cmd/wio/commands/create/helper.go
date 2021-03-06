package create

import (
    "wio/cmd/wio/log"
    "wio/cmd/wio/utils/template"
    "strings"
    "path/filepath"
    "wio/cmd/wio/utils/io"
    "wio/cmd/wio/utils"
    "os"
)

func (info createInfo) fillReadMe(queue *log.Queue, readmeFile string) error {
    log.Verb(queue, "filling README file ... ")
    if err := template.IOReplace(readmeFile, map[string]string{
        "PLATFORM":        info.platform,
        "FRAMEWORK":       info.framework,
        "BOARD":           info.board,
        "PROJECT_NAME":    info.name,
        "PROJECT_VERSION": "0.0.1",
    }); err != nil {
        log.WriteFailure(queue, log.VERB)
        return err
    }
    log.WriteSuccess(queue, log.VERB)
    return nil
}

func (info createInfo) toLowerCase() {
    info.projectType = strings.ToLower(info.projectType)
    info.platform = strings.ToLower(info.platform)
    info.framework = strings.ToLower(info.framework)
    info.board = strings.ToLower(info.board)
}

func (info *createInfo) generateConstraints() (map[string]bool, map[string]bool) {
    context := info.context
    dirConstraints := map[string]bool{
        "tests":          false,
        "no-header-only": !context.Bool("header-only"),
    }
    fileConstraints := map[string]bool{
        "ide=clion":      false,
        "extra":          !context.Bool("no-extras"),
        "example":        context.Bool("create-example"),
        "no-header-only": !context.Bool("no-header-only"),
    }
    return dirConstraints, fileConstraints
}

// This uses a structure.json file and creates a project structure based on that. It takes in consideration
// all the constrains and copies files. This should be used for creating project for any type of app/pkg
func copyProjectAssets(queue *log.Queue, info *createInfo, data *StructureTypeData) error {
    dirConstraints, fileConstraints := info.generateConstraints()
    for _, path := range data.Paths {
        directoryPath := filepath.Clean(info.directory + io.Sep + path.Entry)
        skipDir := false
        log.Verbln(queue, "copying assets to directory: %s", directoryPath)
        // handle directory constraints
        for _, constraint := range path.Constraints {
            _, exists := dirConstraints[constraint]
            if exists && !dirConstraints[constraint] {
                log.Verbln(queue, "constraint not specified and hence skipping this directory")
                skipDir = true
                break
            }
        }
        if skipDir {
            continue
        }

        if !utils.PathExists(directoryPath) {
            if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
                return err
            }
            log.Verbln(queue, "created directory: %s", directoryPath)
        }

        log.Verbln(queue, "copying asset files for directory: %s", directoryPath)
        for _, file := range path.Files {
            toPath := filepath.Clean(directoryPath + io.Sep + file.To)
            skipFile := false
            // handle file constraints
            for _, constraint := range file.Constraints {
                _, exists := fileConstraints[constraint]
                if exists && !fileConstraints[constraint] {
                    log.Verbln(queue, "constraint not specified and hence skipping this file")
                    skipFile = true
                    break
                }
            }
            if skipFile {
                continue
            }

            // handle updates
            if !file.Update && info.updateOnly {
                log.Verbln(queue, "project is not updating, hence skipping update for path: %s", toPath)
                continue
            }

            // copy assets
            if err := io.AssetIO.CopyFile(file.From, toPath, file.Override); err != nil {
                return err
            } else {
                log.Verbln(queue, `copied asset file "%s" TO: %s: `, filepath.Base(file.From), toPath)
            }
        }
    }
    return nil
}
