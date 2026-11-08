#!/usr/bin/env python3
"""Audit CUDA runtime dependencies without copying large toolkit libraries."""
from __future__ import annotations
import argparse, os, re, subprocess
from pathlib import Path
REQUIRED=("libcudart.so.13","libcublas.so.13")
PREFIXES=("libcudart.so","libcublas.so","libcublasLt.so","libnvJitLink.so","libnvrtc.so","libnvrtc-builtins.so","libcurand.so","libcusparse.so","libcusolver.so","libcupti.so","libnvToolsExt.so","libculibos.so")
def main():
 p=argparse.ArgumentParser(); p.add_argument("--root",type=Path,required=True); p.add_argument("--cuda-lib-dir",type=Path,required=True); a=p.parse_args()
 env=os.environ.copy(); env["LD_LIBRARY_PATH"]=str(a.cuda_lib_dir.resolve())+os.pathsep+env.get("LD_LIBRARY_PATH","")
 root=a.root.resolve(); targets=[root/"bin/llama-server",*sorted((root/"bin").glob("*.so*"))]; found={}
 for target in targets:
  if not target.is_file(): continue
  r=subprocess.run(["ldd",str(target)],text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,env=env,timeout=45)
  if r.returncode: raise RuntimeError(f"ldd failed for {target}:\\n{r.stdout}")
  for line in r.stdout.splitlines():
   m=re.match(r"\\s*(\\S+)\\s+=>\\s+(/\\S+)",line)
   if m and m.group(1).startswith(PREFIXES): found[m.group(1)]=m.group(2)
 missing=[x for x in REQUIRED if x not in found]
 if missing: raise RuntimeError("Required CUDA runtime dependencies unresolved: "+", ".join(missing))
 print("CUDA_RUNTIME_AUDIT=PASS")
 for k,v in sorted(found.items()): print(f"{k} => {v}")
 print("CUDA_DRIVER_LIBRARY=HOST_PROVIDED (libcuda.so.1)")
 print("CUDA_RUNTIME_POLICY=COLAB_PYTORCH_DISTRIBUTION")
if __name__=="__main__": main()
