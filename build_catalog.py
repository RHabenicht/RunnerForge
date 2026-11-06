#!/usr/bin/env python3
"""Convert a user-owned NSRCG 2005 installation to typed, filtered JSON.
Original data is not fetched from third-party repositories.
The sr3data project informed the analysis of the original pipe/tree structures;
this implementation is independent and carries source metadata and exclusions.
"""
from __future__ import annotations
import argparse, json, re, collections, pathlib, hashlib

REMOVED = {'sr1','sr2','tss','cb1','cb2','cb3','cb4','cp','pw','n/sl','nagee','nagee7','nagee8','nagee9','adh','bjf','ray'}
# Only explicitly approved legacy source codes may enter the catalog.
# Original r3 records require individual verification against Revised.
APPROVED = {'sr3', 'src', 'cc', 'mits', 'mm', 'mat', 'sota', 'sta2', 'twl', 'yotc'}
FILES = {'gear':'GEAR.DAT','cyberware':'cyber.dat','bioware':'bioware.dat','adept_powers':'adept.dat','spells':'SPELLS.DAT','vehicles':'vehicles.dat','cyberdecks':'DECK.dat','edges_flaws':'EDGE.DAT','magic_gear':'MAGEGEAR.DAT','contacts':'contacts.dat','skills':'SKILLS.DAT','totems':'TOTEMS.DAT'}
TYPE_LINE=re.compile(r'^0-(\d+)\|')
ITEM_LINE=re.compile(r'^(\d+)-\*\s*(.*?)\s+(\d+)$')
SOURCE_REF=re.compile(r'\b([a-z][a-z0-9/]+)\.(\d+|[x?]{2,3}|[a-z]\d+)\b',re.I)

def load_books(base):
    books=[]
    for l in (base/'books.dat').read_text(encoding='cp1252').splitlines():
        if not l.strip() or l.startswith('!') or ';' not in l:continue
        n,abbr=l.split(';',1)
        d=n.strip().startswith('*')
        name=n.strip().removeprefix('*').strip(); abbr=abbr.strip().lower()
        books.append(dict(name=name,code=abbr,default=d,allowed=abbr in APPROVED))
    return books

def refs(fields,known):
    alltext=' '.join(str(v) for v in fields)
    # Avoid treating numbers, measurements or snippets as sourcebook names.
    return sorted(set(x.lower() for x,n in SOURCE_REF.findall(alltext) ))

def valid(rec, allowed):
    rr=rec['sources']
    return bool(rr) and all(r in allowed for r in rr)

def parse_generic(path,kind,known):
    types={};branch=[];out=[];warnings=[]
    for lineno,line in enumerate(path.read_text(encoding='cp1252',errors='replace').splitlines(),1):
        line=line.strip('\ufeff\r\n')
        if not line or line.lstrip().startswith('!'):continue
        t=TYPE_LINE.match(line)
        if t:
            ss=[x.strip() for x in line.rstrip('|').split('|')]
            if len(ss)<3:continue
            ty=int(t.group(1));names=ss[3:]
            types[ty]=(ss[1],names)
            continue
        m=ITEM_LINE.match(line.split('|',1)[0])
        if m and '|' in line:
            depth=int(m.group(1));name=m.group(2).strip()
            ty=int(m.group(3));fields=[x.strip() for x in line.split('|')[1:]]
            if ty not in types:
                warnings.append(f'{kind}:{lineno}: unknown type {ty}');continue
            tn,fn=types[ty]
            attrs={key:fields[ix] if ix<len(fields) else '' for ix,key in enumerate(fn)}
            if len(fields)>len(fn):attrs['_extra_fields']=fields[len(fn):]
            category=' / '.join(branch[:depth-1])
            r=refs(attrs.values(),known)
            out.append(dict(name=name,kind=kind,group=tn,category=category,fields=attrs,sources=r,origin=f'{path.name}:{lineno}'))
            continue
        # Branch begins with N-; for book-specific data as well as nested weapon groups.
        h=re.match(r'^([1-9])-([^|]+)$',line)
        if h:
            depth=int(h.group(1))
            branch=branch[:depth-1]+[h.group(2).strip()]
            continue
        # Some files have e.g '* Item', handled via fallback for fewer cases.
        if not line.startswith(('!','#')):warnings.append(f'{kind}:{lineno}: unsupported line {line[:45]}')
    return out,warnings

def parse_skills(path,known):
    out=[]; cls='';current=None
    for lineno,line in enumerate(path.read_text(encoding='cp1252',errors='replace').splitlines(),1):
        if not line.strip() or line.startswith('!'):continue
        if line.startswith('#'):
            cls=line.lstrip('#').strip();continue
        if line.startswith(' '):
            if current and ('|' in line or line.strip().endswith('->')):
                current['fields'].setdefault('specializations',[]).append(line.strip())
            continue
        pts=line.split('|'); nm=pts[0].strip()
        if not nm or '(' not in nm:continue
        name,attrib=nm.rsplit('(',1);attrib=attrib.rstrip(')')
        source=pts[1].strip() if len(pts)>1 else ''
        rr=refs([source],known)
        current=dict(name=name,kind='skills',group=cls,category=attrib,fields={'attribute':attrib,'page_reference':source,'notes':pts[2] if len(pts)>2 else '', 'specializations':[]},sources=rr,origin=f'{path.name}:{lineno}')
        out.append(current)
    return out,[]

def parse_contacts(path,known):
    out=[];category=''
    for lineno,line in enumerate(path.read_text(encoding='cp1252',errors='replace').splitlines(),1):
        raw=line.strip()
        if not raw or raw.startswith(('!','0-')):continue
        h=re.match(r'^1-(.*)$',raw)
        if h:category=h.group(1).strip();continue
        item=re.match(r'^2-\*\s*(.*)$',raw)
        if item:
            name=item.group(1).strip()
            if name:
                out.append(dict(name=name,kind='contacts',group='Contacts',category=category,fields={},sources=[],origin=f'{path.name}:{lineno}'))
    return out,[]

def parse_totems(path,known):
    out=[];category='';current=None
    for lineno,line in enumerate(path.read_text(encoding='cp1252',errors='replace').splitlines(),1):
        t=line.strip()
        if not t or t.startswith('!'):continue
        if t.startswith('\\'):
            category=t[1:].strip();continue
        if t.startswith('*'):
            parts=t[1:].split('|');name=parts[0].strip();page=parts[1].strip() if len(parts)>1 else ''
            current=dict(name=name,kind='totems',group='Totems',category=category,fields={'page_reference':page,'type':parts[2].strip() if len(parts)>2 else '','description':''},sources=refs([page],known),origin=f'{path.name}:{lineno}')
            out.append(current)
        elif current:current['fields']['description']+=' '+t
    return out,[]

def write_catalog(books, records, warnings, outpath, quarantine_path):
    """Partition records without renumbering existing character references."""
    allowed = {book['code'] for book in books if book['code'] in APPROVED}
    for book in books:
        book['allowed'] = book['code'] in allowed
    ids = [record['id'] for record in records]
    if len(ids) != len(set(ids)) or any(value <= 0 for value in ids):
        raise ValueError('Catalog IDs must be unique positive integers')
    retained = []
    quarantine = []
    for record in sorted(records, key=lambda item: item['id']):
        # Inspect complete fields, including descriptions and specializations.
        citations = refs(record['fields'].values(), allowed)
        record['sources'] = sorted(set(record['sources']) | set(citations))
        if valid(record, allowed):
            retained.append(record)
        else:
            quarantine.append(record)
    statistics = {
        kind: {
            'read': sum(record['kind'] == kind for record in records),
            'retained': sum(record['kind'] == kind for record in retained),
            'excluded': sum(record['kind'] == kind for record in quarantine),
        }
        for kind in FILES
    }
    payload = dict(
        schema='nsrcgii.catalog.v1',
        source='Local user-supplied NSRCG 2005 installer',
        edition='SR3', books=books, entries=retained,
        statistics=statistics, parser_warnings=warnings,
    )
    for path, content in [(outpath, payload), (quarantine_path, quarantine)]:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            json.dumps(content, ensure_ascii=False, separators=(',', ':')),
            encoding='utf-8',
        )
    return payload


def build(base, outpath, quarantine_path):
    books = load_books(base)
    known = {book['code'] for book in books}
    entries = []
    warnings = []
    for kind, name in FILES.items():
        path = base / name
        if not path.exists():
            raise FileNotFoundError(path)
        parsers = {
            'skills': parse_skills,
            'contacts': parse_contacts,
            'totems': parse_totems,
        }
        if kind in parsers:
            records, issues = parsers[kind](path, known)
        else:
            records, issues = parse_generic(path, kind, known)
        entries.extend(records)
        warnings.extend(issues)
    for index, record in enumerate(entries, 1):
        record['id'] = index
    return write_catalog(books, entries, warnings, outpath, quarantine_path)


if __name__ == '__main__':
    root = pathlib.Path(__file__).parent
    parser = argparse.ArgumentParser()
    inputs = parser.add_mutually_exclusive_group(required=True)
    inputs.add_argument('--original', type=pathlib.Path)
    inputs.add_argument('--catalog', type=pathlib.Path,
                        help='Recheck an existing catalog together with --quarantine')
    parser.add_argument('--output', type=pathlib.Path,
                        default=root / 'go/srcapp/catalog.json')
    parser.add_argument('--quarantine', type=pathlib.Path,
                        default=root / 'go/data/quarantine.json')
    args = parser.parse_args()
    if args.catalog:
        existing = json.loads(args.catalog.read_text(encoding='utf-8'))
        excluded = json.loads(args.quarantine.read_text(encoding='utf-8'))
        obj = write_catalog(
            existing['books'], existing['entries'] + excluded,
            existing['parser_warnings'], args.output, args.quarantine,
        )
    else:
        obj = build(args.original, args.output, args.quarantine)
    print('Retained', len(obj['entries']))
    print('Quarantined', sum(row['excluded'] for row in obj['statistics'].values()))
