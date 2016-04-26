package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/steveoc64/go-cmms/shared"
)

type UtilRPC struct{}

// Do a simple database backup
func (u *UtilRPC) Backup(channel int, result *string) error {
	start := time.Now()

	conn := Connections.Get(channel)
	*result = ""

	if conn.UserRole == "Admin" {
		out, err := exec.Command("../scripts/cmms-backup.sh").Output()
		if err != nil {
			log.Println(err)
			*result = err.Error()
			return nil
		}
		*result = string(out)
	}

	logger(start, "Util.Backup",
		fmt.Sprintf("Channel %d, User %d %s %s",
			channel, conn.UserID, conn.Username, conn.UserRole),
		*result)

	return nil
}

// Run a top command and return the results
func (u *UtilRPC) Top(channel int, result *string) error {
	start := time.Now()

	conn := Connections.Get(channel)
	*result = ""

	if conn.UserRole == "Admin" {
		out, err := exec.Command("../scripts/top.sh").Output()
		if err != nil {
			log.Println(err)
			*result = err.Error()
			return nil
		}
		*result = string(out)
	}

	logger(start, "Util.Top",
		fmt.Sprintf("Channel %d, User %d %s %s",
			channel, conn.UserID, conn.Username, conn.UserRole),
		"top")

	return nil
}

// Tail the output of the server process log file, and return the result
func (u *UtilRPC) Logs(channel int, result *string) error {
	start := time.Now()

	conn := Connections.Get(channel)
	*result = ""

	if conn.UserRole == "Admin" {
		out, err := exec.Command("../scripts/logs.sh").Output()
		if err != nil {
			log.Println(err)
			*result = err.Error()
			return nil
		}
		*result = string(out)
	}

	logger(start, "Util.Logs",
		fmt.Sprintf("Channel %d, User %d %s %s",
			channel, conn.UserID, conn.Username, conn.UserRole),
		"logs")

	return nil
}

// Patch up tha machine PartClass - Steve only
func (u *UtilRPC) Machine(channel int, result *string) error {
	start := time.Now()

	conn := Connections.Get(channel)
	*result = ""

	if conn.UserRole == "Admin" && conn.Username == "steve" {

		// For each machine, set the class the same as the name, and report any errors
		r := "Processing Machines\n"

		machines := []shared.Machine{}
		partClass := shared.PartClass{}

		DB.SQL(`select 
			m.*,s.name as site_name
			from machine m
			left join site s on s.id=m.site_id
			where part_class=0
			order by m.id`).
			QueryStructs(&machines)

		patched := 0
		for _, m := range machines {
			siteName := ""
			if m.SiteName != nil {
				siteName = *m.SiteName
			}
			r += fmt.Sprintf("Machine %d: %s (%s)", m.ID, m.Name, siteName)

			err := DB.SQL(`select * from part_class where name=$1`, m.Name).QueryStruct(&partClass)
			if err != nil {
				r += fmt.Sprintf("\n    !! No Matching Part Class !!\n")
				continue
			}

			if partClass.ID == 0 {
				r += fmt.Sprintf("\n    !! No Matching Part Class !!\n")
			} else {
				r += fmt.Sprintf(" = PartClass %d: %s\n", partClass.ID, partClass.Name)
				DB.SQL(`update machine set part_class=$1 where id=$2`, partClass.ID, m.ID).Exec()
				patched++
			}
		}

		r += fmt.Sprintf("\nPatched %d of %d Machines\n", patched, len(machines))
		*result = r
	}

	logger(start, "Util.Machine",
		fmt.Sprintf("Channel %d, User %d %s %s",
			channel, conn.UserID, conn.Username, conn.UserRole),
		*result)

	return nil
}

// Patch up the Parts PartClass field - Steve only
func (u *UtilRPC) Parts(channel int, result *string) error {
	start := time.Now()

	conn := Connections.Get(channel)

	if conn.UserRole == "Admin" && conn.Username == "steve" {
		// For each part, get the 1st component that its associated
		// with (under the old scheme), and from there get the machine.
		//
		// The machine then tells us which partclass to use
		r := "Processing Parts\n"

		parts := []shared.Part{}

		DB.SQL(`select * from part
			where class=0
			order by id`).
			QueryStructs(&parts)

		patched := 0
		for _, p := range parts {
			r += fmt.Sprintf("Part %d: %s", p.ID, p.Name)

			// Get the first associated component
			classID := 0
			err := DB.SQL(`select 
				m.part_class
				from component_part x
				left join component c on c.id=x.component_id
				left join machine m on m.id=c.machine_id
				where x.part_id=$1
				limit 1`, p.ID).QueryScalar(&classID)

			if err != nil {
				r += fmt.Sprintf("\n    !! No can find the partclass\n")
				continue
			}

			if classID == 0 {
			} else {
				r += fmt.Sprintf(" = PartClass %d\n", classID)
				DB.SQL(`update part set class=$1 where id=$2`, classID, p.ID).Exec()
				patched++
			}
		}

		r += fmt.Sprintf("\nPatched %d of %d Parts\n", patched, len(parts))
		*result = r
	}

	logger(start, "Util.Parts",
		fmt.Sprintf("Channel %d, User %d %s %s",
			channel, conn.UserID, conn.Username, conn.UserRole),
		*result)

	return nil
}