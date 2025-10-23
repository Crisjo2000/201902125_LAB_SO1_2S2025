#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/string.h> 
#include <linux/init.h>
#include <linux/proc_fs.h> 
#include <linux/seq_file.h> 
#include <linux/mm.h> 
#include <linux/sched.h> 
#include <linux/timer.h> 
#include <linux/jiffies.h> 
#include <linux/uaccess.h>
#include <linux/tty.h>
#include <linux/sched/signal.h>
#include <linux/fs.h>        
#include <linux/slab.h>      
#include <linux/sched/mm.h>
#include <linux/binfmts.h>
#include <linux/timekeeping.h>
#include <linux/sysinfo.h>

MODULE_LICENSE("GPL");
MODULE_AUTHOR("CHRISTIAN HERNANDEZ - 201902125");
MODULE_DESCRIPTION("Modulo para leer informacion de memoria y CPU en JSON");
MODULE_VERSION("1.0");

#define PROC_NAME "sysinfo_so1_201902125"
#define MAX_CMDLINE_LENGTH 256

// Función para obtener la línea de comandos
static char *get_process_cmdline(struct task_struct *task) {
    struct mm_struct *mm;
    char *cmdline, *p;
    unsigned long arg_start, arg_end, env_start;
    int i, len;

    cmdline = kmalloc(MAX_CMDLINE_LENGTH, GFP_KERNEL);
    if (!cmdline)
        return NULL;

    mm = get_task_mm(task);
    if (!mm) {
        kfree(cmdline);
        return NULL;
    }

    down_read(&mm->mmap_lock);
    arg_start = mm->arg_start;
    arg_end = mm->arg_end;
    env_start = mm->env_start;
    up_read(&mm->mmap_lock);

    len = arg_end - arg_start;
    if (len > MAX_CMDLINE_LENGTH - 1)
        len = MAX_CMDLINE_LENGTH - 1;

    if (access_process_vm(task, arg_start, cmdline, len, 0) != len) {
        mmput(mm);
        kfree(cmdline);
        return NULL;
    }

    cmdline[len] = '\0';
    p = cmdline;
    for (i = 0; i < len; i++)
        if (p[i] == '\0')
            p[i] = ' ';

    mmput(mm);
    return cmdline;
}

// Función para mostrar todos los procesos

static int sysinfo_show(struct seq_file *m, void *v) {
    struct sysinfo si;
    struct task_struct *task;
    int first_process = 1;
    int process_count = 0;

    si_meminfo(&si);

    // Contador de procesos
    for_each_process(task) {
        process_count++;
    }

    seq_printf(m, "{\n");
    seq_printf(m, "  \"Totalram\": %lu,\n", si.totalram * (PAGE_SIZE / 1024));
    seq_printf(m, "  \"Freeram\": %lu,\n", si.freeram * (PAGE_SIZE / 1024));
    seq_printf(m, "  \"Procs\": %d,\n", process_count);
    seq_printf(m, "  \"Processes\": [\n");

    for_each_process(task) {
        unsigned long vsz = 0;
        unsigned long rss = 0;
        unsigned long totalram = si.totalram * (PAGE_SIZE / 1024);
        unsigned long mem_usage = 0;
        unsigned long cpu_usage = 0;
        char *cmdline = NULL;

        if (task->mm) {
            vsz = task->mm->total_vm * (PAGE_SIZE / 1024);
            rss = get_mm_rss(task->mm) * (PAGE_SIZE / 1024);
            if (totalram > 0 && rss > 0) {
                mem_usage = (rss * 1000) / totalram;
            }
        }

        unsigned long total_time = task->utime + task->stime;
        cpu_usage = total_time / 1000000;

        cmdline = get_process_cmdline(task);

        if (!first_process) {
            seq_printf(m, ",\n");
        } else {
            first_process = 0;
        }

        seq_printf(m, "     {\n");
        seq_printf(m, "        \"PID\": %d,\n", task->pid);
        seq_printf(m, "        \"Name\": \"%s\",\n", task->comm);
        seq_printf(m, "        \"Cmdline\": \"%s\",\n", cmdline ? cmdline : "N/A");
        seq_printf(m, "        \"vsz\": %lu,\n", vsz);
        seq_printf(m, "        \"rss\": %lu,\n", rss);
        seq_printf(m, "        \"Memory_Usage\": %lu.%lu,\n", mem_usage / 10, mem_usage % 10);
        seq_printf(m, "        \"CPU_Usage\": %lu\n", cpu_usage);
        seq_printf(m, "     }");

        if (cmdline) {
            kfree(cmdline);
        }
    }

    seq_printf(m, "\n   ]\n}\n");
    return 0;
}

// Operaciones del archivo /proc
static int sysinfo_open(struct inode *inode, struct file *file) {
    return single_open(file, sysinfo_show, NULL);
}

static const struct proc_ops sysinfo_ops = {
    .proc_open = sysinfo_open,
    .proc_read = seq_read,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};

static int __init sysinfo_init(void) {
    proc_create(PROC_NAME, 0, NULL, &sysinfo_ops);
    printk(KERN_INFO "Módulo sysinfo cargado: /proc/%s\n", PROC_NAME);
    return 0;
}

static void __exit sysinfo_exit(void) {
    remove_proc_entry(PROC_NAME, NULL);
    printk(KERN_INFO "Módulo sysinfo descargado\n");
}

module_init(sysinfo_init);
module_exit(sysinfo_exit);