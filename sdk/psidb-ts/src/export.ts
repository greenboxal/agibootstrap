import "reflect-metadata";
import { reflect } from 'typescript-rtti';
import {ModuleInterface} from "./index";

const ModuleInterfaceType = reflect<ModuleInterface>();
const ModuleInterfaceIface = ModuleInterfaceType.as('interface')

console.log(ModuleInterfaceType.kind)
console.log(Object.getOwnPropertyNames(ModuleInterfaceIface.reflectedInterface.class.prototype))