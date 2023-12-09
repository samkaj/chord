## Test commands

**Create Ring**
```bash
./chord.sh -a 0.0.0.0:3000 -tcp 5000 -ts 5000 -ff 5000 -r 4
```

**Join Ring 1**
```bash
./chord.sh -a 0.0.0.0:3001 -j 0.0.0.0:3000 -tcp 5000 -ts 5000 -ff 5000 -r 4
```

**Join ring 2**
```bash
./chord.sh -a 0.0.0.0:3002 -j 0.0.0.0:3000 -tcp 5000 -ts 5000 -ff 5000 -r 4
```

**Join ring 3**
```bash
./chord.sh -a 0.0.0.0:3003 -j 0.0.0.0:3000 -tcp 5000 -ts 5000 -ff 5000 -r 4
```

